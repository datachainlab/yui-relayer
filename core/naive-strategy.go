package core

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	retry "github.com/avast/retry-go"
	sdk "github.com/cosmos/cosmos-sdk/types"
	chantypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	"golang.org/x/sync/errgroup"

	"github.com/hyperledger-labs/yui-relayer/utils"
)

// NaiveStrategy is an implementation of Strategy.
type NaiveStrategy struct {
	Ordered      bool
	MaxTxSize    uint64 // maximum permitted size of the msgs in a bundled relay transaction
	MaxMsgLength uint64 // maximum amount of messages in a bundled relay transaction
}

var _ StrategyI = (*NaiveStrategy)(nil)

const MaxMsgLength = 400

func NewNaiveStrategy() *NaiveStrategy {
	return &NaiveStrategy{}
}

// GetType implements Strategy
func (st NaiveStrategy) GetType() string {
	return "naive"
}

func (st NaiveStrategy) SetupRelay(ctx context.Context, src, dst *ProvableChain) error {
	defer utils.Track(time.Now(), "SetupRelay()", nil)

	if err := src.SetupForRelay(ctx); err != nil {
		return err
	}
	if err := dst.SetupForRelay(ctx); err != nil {
		return err
	}
	return nil
}

func (st NaiveStrategy) UnrelayedSequences(src, dst *ProvableChain, sh SyncHeadersI) (*RelaySequences, error) {
	defer utils.Track(time.Now(), "UnrelayedSequences()", nil)

	var (
		eg           = new(errgroup.Group)
		srcPacketSeq = []uint64{}
		dstPacketSeq = []uint64{}
		err          error
		rs           = &RelaySequences{Src: []uint64{}, Dst: []uint64{}}
	)

	eg.Go(func() error {
		var res *chantypes.QueryPacketCommitmentsResponse
		if err = retry.Do(func() error {
			// Query the packet commitment
			res, err = src.QueryPacketCommitments(0, 1000, sh.GetQueryableHeight(src.ChainID()))
			switch {
			case err != nil:
				return err
			case res == nil:
				return fmt.Errorf("No error on QueryPacketCommitments for %s, however response is nil", src.ChainID())
			default:
				return nil
			}
		}, rtyAtt, rtyDel, rtyErr, retry.OnRetry(func(n uint, err error) {
			log.Println(fmt.Sprintf("- [%s]@{%d} - try(%d/%d) query packet commitments: %s", src.ChainID(), sh.GetQueryableHeight(src.ChainID()), n+1, rtyAttNum, err))
		})); err != nil {
			return err
		}
		for _, pc := range res.Commitments {
			srcPacketSeq = append(srcPacketSeq, pc.Sequence)
		}
		return nil
	})

	eg.Go(func() error {
		var res *chantypes.QueryPacketCommitmentsResponse
		if err = retry.Do(func() error {
			res, err = dst.QueryPacketCommitments(0, 1000, sh.GetQueryableHeight(dst.ChainID()))
			switch {
			case err != nil:
				return err
			case res == nil:
				return fmt.Errorf("No error on QueryPacketCommitments for %s, however response is nil", dst.ChainID())
			default:
				return nil
			}
		}, rtyAtt, rtyDel, rtyErr, retry.OnRetry(func(n uint, err error) {
			log.Println(fmt.Sprintf("- [%s]@{%d} - try(%d/%d) query packet commitments: %s", dst.ChainID(), sh.GetQueryableHeight(dst.ChainID()), n+1, rtyAttNum, err))
		})); err != nil {
			return err
		}
		for _, pc := range res.Commitments {
			dstPacketSeq = append(dstPacketSeq, pc.Sequence)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	eg.Go(func() error {
		// Query all packets sent by src that have been received by dst
		src, err := dst.QueryUnrecievedPackets(sh.GetQueryableHeight(dst.ChainID()), srcPacketSeq)
		if err != nil {
			return err
		} else if src != nil {
			rs.Src = src
		}
		return nil
	})

	eg.Go(func() error {
		// Query all packets sent by dst that have been received by src
		dst, err := src.QueryUnrecievedPackets(sh.GetQueryableHeight(src.ChainID()), dstPacketSeq)
		if err != nil {
			return err
		} else if dst != nil {
			rs.Dst = dst
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return rs, nil
}

func (st NaiveStrategy) RelayPackets(src, dst *ProvableChain, sp *RelaySequences, sh SyncHeadersI) error {
	defer utils.Track(time.Now(), "RelayPackets()", nil)

	// set the maximum relay transaction constraints
	msgs := &RelayMsgs{
		Src:       []sdk.Msg{},
		Dst:       []sdk.Msg{},
		MaxTxSize: st.MaxTxSize,
		//MaxMsgLength: st.MaxMsgLength,
		MaxMsgLength: MaxMsgLength,
	}
	addr, err := dst.GetAddress()
	if err != nil {
		return err
	}
	msgs.Dst, err = relayPacketsConcurrent(src, sp.Src, sh, addr)
	//msgs.Dst, err = relayPacketsInBulk(src, sp.Src, sh, addr)
	if err != nil {
		return err
	}
	addr, err = src.GetAddress()
	if err != nil {
		return err
	}
	msgs.Src, err = relayPacketsConcurrent(dst, sp.Dst, sh, addr)
	//msgs.Src, err = relayPacketsInBulk(dst, sp.Dst, sh, addr)
	if err != nil {
		return err
	}
	if !msgs.Ready() {
		log.Println(fmt.Sprintf("- No packets to relay between [%s]port{%s} and [%s]port{%s}",
			src.ChainID(), src.Path().PortID, dst.ChainID(), dst.Path().PortID))
		//log.Printf("sp.Src: %d, sp.Dst: %d, msgs.Dst: %d, msgs.Src: %d", len(sp.Src), len(sp.Dst), len(msgs.Dst), len(msgs.Src))
		return nil
	}

	// Prepend non-empty msg lists with UpdateClient
	if len(msgs.Dst) != 0 {
		// Sending an update from src to dst
		h, err := sh.GetHeader(src, dst)
		if err != nil {
			return err
		}
		addr, err := dst.GetAddress()
		if err != nil {
			return err
		}
		if h != nil {
			msgs.Dst = append([]sdk.Msg{dst.Path().UpdateClient(h, addr)}, msgs.Dst...)
		}
	}

	if len(msgs.Src) != 0 {
		h, err := sh.GetHeader(dst, src)
		if err != nil {
			return err
		}
		addr, err := src.GetAddress()
		if err != nil {
			return err
		}
		if h != nil {
			msgs.Src = append([]sdk.Msg{src.Path().UpdateClient(h, addr)}, msgs.Src...)
		}
	}

	// send messages to their respective chains
	if msgs.Send(src, dst); msgs.Success() {
		if len(msgs.Dst) > 1 {
			logPacketsRelayed(dst, src, len(msgs.Dst)-1)
		}
		if len(msgs.Src) > 1 {
			logPacketsRelayed(src, dst, len(msgs.Src)-1)
		}
	}

	return nil
}

func (st NaiveStrategy) UnrelayedAcknowledgements(src, dst *ProvableChain, sh SyncHeadersI) (*RelaySequences, error) {
	defer utils.Track(time.Now(), "UnrelayedAcknowledgements()", nil)

	var (
		eg           = new(errgroup.Group)
		srcPacketSeq = []uint64{}
		dstPacketSeq = []uint64{}
		err          error
		rs           = &RelaySequences{Src: []uint64{}, Dst: []uint64{}}
	)

	eg.Go(func() error {
		var res *chantypes.QueryPacketAcknowledgementsResponse
		if err = retry.Do(func() error {
			// Query the packet commitment
			res, err = src.QueryPacketAcknowledgementCommitments(0, 1000, sh.GetQueryableHeight(src.ChainID()))
			switch {
			case err != nil:
				return err
			case res == nil:
				return fmt.Errorf("No error on QueryPacketUnrelayedAcknowledgements for %s, however response is nil", src.ChainID())
			default:
				return nil
			}
		}, rtyAtt, rtyDel, rtyErr, retry.OnRetry(func(n uint, err error) {
			log.Println((fmt.Sprintf("- [%s]@{%d} - try(%d/%d) query packet acknowledgements: %s", src.ChainID(), sh.GetQueryableHeight(src.ChainID()), n+1, rtyAttNum, err)))
			sh.Updates(src, dst)
		})); err != nil {
			return err
		}
		for _, pc := range res.Acknowledgements {
			srcPacketSeq = append(srcPacketSeq, pc.Sequence)
		}
		return nil
	})

	eg.Go(func() error {
		var res *chantypes.QueryPacketAcknowledgementsResponse
		if err = retry.Do(func() error {
			res, err = dst.QueryPacketAcknowledgementCommitments(0, 1000, sh.GetQueryableHeight(dst.ChainID()))
			switch {
			case err != nil:
				return err
			case res == nil:
				return fmt.Errorf("No error on QueryPacketUnrelayedAcknowledgements for %s, however response is nil", dst.ChainID())
			default:
				return nil
			}
		}, rtyAtt, rtyDel, rtyErr, retry.OnRetry(func(n uint, err error) {
			log.Println((fmt.Sprintf("- [%s]@{%d} - try(%d/%d) query packet acknowledgements: %s", dst.ChainID(), sh.GetQueryableHeight(dst.ChainID()), n+1, rtyAttNum, err)))
			sh.Updates(src, dst)
		})); err != nil {
			return err
		}
		for _, pc := range res.Acknowledgements {
			dstPacketSeq = append(dstPacketSeq, pc.Sequence)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	eg.Go(func() error {
		// Query all packets sent by src that have been received by dst
		src, err := dst.QueryUnrecievedAcknowledgements(sh.GetQueryableHeight(dst.ChainID()), srcPacketSeq)
		// return err
		if err != nil {
			return err
		} else if src != nil {
			rs.Src = src
		}
		return nil
	})

	eg.Go(func() error {
		// Query all packets sent by dst that have been received by src
		dst, err := src.QueryUnrecievedAcknowledgements(sh.GetQueryableHeight(src.ChainID()), dstPacketSeq)
		if err != nil {
			return err
		} else if dst != nil {
			rs.Dst = dst
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return rs, nil
}

// TODO add packet-timeout support
func relayPackets(chain *ProvableChain, seqs []uint64, sh SyncHeadersI, sender sdk.AccAddress) ([]sdk.Msg, error) {
	logData := map[string]string{"seqs": fmt.Sprintf("%d", len(seqs))}
	defer utils.Track(time.Now(), "Prover.relayPackets()", logData)

	var msgs []sdk.Msg
	for _, seq := range seqs {
		p, err := chain.QueryPacket(int64(sh.GetQueryableHeight(chain.ChainID())), seq)
		if err != nil {
			log.Println("failed to QueryPacket:", int64(sh.GetQueryableHeight(chain.ChainID())), seq, err)
			return nil, err
		}
		provableHeight := sh.GetProvableHeight(chain.ChainID())
		res, err := chain.QueryPacketCommitmentWithProof(provableHeight, seq)
		if err != nil {
			log.Println("failed to QueryPacketCommitment:", provableHeight, seq, err)
			return nil, err
		}
		msg := chantypes.NewMsgRecvPacket(*p, res.Proof, res.ProofHeight, sender.String())
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

// Note: This pattern doesn't work due to query's condition limitation
func relayPacketsInBulk(chain *ProvableChain, seqs []uint64, sh SyncHeadersI, sender sdk.AccAddress) ([]sdk.Msg, error) {
	logData := map[string]string{"seqs": fmt.Sprintf("%d", len(seqs))}
	defer utils.Track(time.Now(), "Prover.relayPackets()", logData)

	msgs := make([]sdk.Msg, 0, len(seqs))
	provableHeight := sh.GetProvableHeight(chain.ChainID())
	queryableHeight := sh.GetQueryableHeight(chain.ChainID())

	packets, err := chain.QueryPackets(queryableHeight, seqs)
	if err != nil {
		log.Println("failed to QueryPackets:", queryableHeight, err)
		return nil, err
	}
	//FIXME: concurrency
	for _, seq := range seqs {
		res, err := chain.QueryPacketCommitmentWithProof(provableHeight, seq)
		if err != nil {
			log.Println("failed to QueryPacketCommitment:", provableHeight, seq, err)
			return nil, err
		}
		if packet, ok := packets[seq]; ok {
			msg := chantypes.NewMsgRecvPacket(*packet, res.Proof, res.ProofHeight, sender.String())
			msgs = append(msgs, msg)
		}
	}
	return msgs, nil
}

// TODO add packet-timeout support
// TODO: switch to concurency
func relayPacketsConcurrent(chain *ProvableChain, seqs []uint64, sh SyncHeadersI, sender sdk.AccAddress) ([]sdk.Msg, error) {
	logData := map[string]string{"seqs": fmt.Sprintf("%d", len(seqs))}
	defer utils.Track(time.Now(), "relayPacketsConcurrent()", logData)

	if len(seqs) == 0 {
		return []sdk.Msg{}, nil
	}

	msgs := make([]sdk.Msg, 0, len(seqs))

	provableHeight := sh.GetProvableHeight(chain.ChainID())
	queryableHeight := sh.GetQueryableHeight(chain.ChainID())

	wg := &sync.WaitGroup{}

	// TODO: This number must be tweaked
	// The following error occurred at 100 concurrency
	// failed to QueryPacket: 113 103 post failed: Post "http://localhost:26657":
	//  context deadline exceeded (Client.Timeout exceeded while awaiting headers)
	chSemaphore := make(chan struct{}, 30)

	chMsg := make(chan sdk.Msg)
	chErr := make(chan error)

	wg.Add(1)
	go func() {
		for i, seq := range seqs {
			wg.Add(1)
			chSemaphore <- struct{}{}

			go func(idx int, sq uint64) {
				defer func() {
					<-chSemaphore
					if idx == 0 {
						// to wait goroutine for closing channel
						wg.Done()
					}
					wg.Done()
				}()
				// concurrent logic
				p, err := chain.QueryPacket(queryableHeight, sq)
				if err != nil {
					if strings.Contains(err.Error(), "Client.Timeout") {
						// TODO: skip
						log.Println("skip chain.QueryPacket() due to timeout: ", err)
						return
					}
					log.Println("failed to QueryPacket:", queryableHeight, sq, err)
					chErr <- err
					return
				}
				res, err := chain.QueryPacketCommitmentWithProof(provableHeight, sq)
				if err != nil {
					if strings.Contains(err.Error(), "Client.Timeout") {
						// TODO: skip
						log.Println("skip chain.QueryPacketCommitmentWithProof() due to timeout: ", err)
						return
					}
					log.Println("failed to QueryPacketCommitment:", provableHeight, sq, err)
					chErr <- err
					return
				}
				chMsg <- chantypes.NewMsgRecvPacket(*p, res.Proof, res.ProofHeight, sender.String())
			}(i, seq)
		}
	}()
	go func() {
		wg.Wait()
		close(chMsg)
		close(chErr)
	}()

	// wait until channel is closed
	for {
		select {
		case msg, ok := <-chMsg:
			if !ok {
				return msgs, nil
			}
			msgs = append(msgs, msg)
		case err, ok := <-chErr:
			if ok {
				return nil, err
			}
		}
	}
}

func logPacketsRelayed(src, dst ChainI, num int) {
	log.Println(fmt.Sprintf("★ Relayed %d packets: [%s]port{%s}->[%s]port{%s}",
		num, dst.ChainID(), dst.Path().PortID, src.ChainID(), src.Path().PortID))
}

func (st NaiveStrategy) RelayAcknowledgements(src, dst *ProvableChain, sp *RelaySequences, sh SyncHeadersI) error {
	defer utils.Track(time.Now(), "RelayAcknowledgements()", nil)

	// set the maximum relay transaction constraints
	msgs := &RelayMsgs{
		Src:       []sdk.Msg{},
		Dst:       []sdk.Msg{},
		MaxTxSize: st.MaxTxSize,
		//MaxMsgLength: st.MaxMsgLength,
		MaxMsgLength: MaxMsgLength,
	}

	addr, err := dst.GetAddress()
	if err != nil {
		return err
	}
	msgs.Dst, err = relayAcksConcurency(src, dst, sp.Src, sh, addr)
	if err != nil {
		return err
	}
	addr, err = src.GetAddress()
	if err != nil {
		return err
	}
	msgs.Src, err = relayAcksConcurency(dst, src, sp.Dst, sh, addr)
	if err != nil {
		return err
	}
	if !msgs.Ready() {
		log.Println(fmt.Sprintf("- No acknowledgements to relay between [%s]port{%s} and [%s]port{%s}",
			src.ChainID(), src.Path().PortID, dst.ChainID(), dst.Path().PortID))
		//log.Printf("sp.Src: %d, sp.Dst: %d, msgs.Dst: %d, msgs.Src: %d", len(sp.Src), len(sp.Dst), len(msgs.Dst), len(msgs.Src))
		return nil
	}

	// Prepend non-empty msg lists with UpdateClient
	if len(msgs.Dst) != 0 {
		h, err := sh.GetHeader(src, dst)
		if err != nil {
			return err
		}
		addr, err := dst.GetAddress()
		if err != nil {
			return err
		}
		if h != nil {
			msgs.Dst = append([]sdk.Msg{dst.Path().UpdateClient(h, addr)}, msgs.Dst...)
		}
	}

	if len(msgs.Src) != 0 {
		h, err := sh.GetHeader(dst, src)
		if err != nil {
			return err
		}
		addr, err := src.GetAddress()
		if err != nil {
			return err
		}
		if h != nil {
			msgs.Src = append([]sdk.Msg{src.Path().UpdateClient(h, addr)}, msgs.Src...)
		}
	}

	// send messages to their respective chains
	if msgs.Send(src, dst); msgs.Success() {
		if len(msgs.Dst) > 1 {
			logPacketsRelayed(dst, src, len(msgs.Dst)-1)
		}
		if len(msgs.Src) > 1 {
			logPacketsRelayed(src, dst, len(msgs.Src)-1)
		}
	}

	return nil
}

func relayAcks(receiverChain, senderChain *ProvableChain, seqs []uint64, sh SyncHeadersI, sender sdk.AccAddress) ([]sdk.Msg, error) {
	logData := map[string]string{"seqs": fmt.Sprintf("%d", len(seqs))}
	defer utils.Track(time.Now(), "Prover.relayAcks()", logData)

	var msgs []sdk.Msg

	for _, seq := range seqs {
		p, err := senderChain.QueryPacket(sh.GetQueryableHeight(senderChain.ChainID()), seq)
		if err != nil {
			return nil, err
		}
		ack, err := receiverChain.QueryPacketAcknowledgement(sh.GetQueryableHeight(receiverChain.ChainID()), seq)
		if err != nil {
			return nil, err
		}
		provableHeight := sh.GetProvableHeight(receiverChain.ChainID())
		res, err := receiverChain.QueryPacketAcknowledgementCommitmentWithProof(provableHeight, seq)
		if err != nil {
			return nil, err
		}

		msg := chantypes.NewMsgAcknowledgement(*p, ack, res.Proof, res.ProofHeight, sender.String())
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

// TODO: switch to concurency
func relayAcksConcurency(receiverChain, senderChain *ProvableChain, seqs []uint64, sh SyncHeadersI, sender sdk.AccAddress) ([]sdk.Msg, error) {
	logData := map[string]string{"seqs": fmt.Sprintf("%d", len(seqs))}
	defer utils.Track(time.Now(), "relayAcksConcurency()", logData)

	if len(seqs) == 0 {
		return []sdk.Msg{}, nil
	}

	msgs := make([]sdk.Msg, 0, len(seqs))
	provableHeight := sh.GetProvableHeight(receiverChain.ChainID())
	senderQueryableHeight := sh.GetQueryableHeight(senderChain.ChainID())
	receiverQueryableHeight := sh.GetQueryableHeight(receiverChain.ChainID())

	wg := &sync.WaitGroup{}

	// TODO: This number must be tweaked
	// The following error occurred at 100 concurrency
	// failed to QueryPacket: 113 103 post failed: Post "http://localhost:26657":
	//  context deadline exceeded (Client.Timeout exceeded while awaiting headers)
	chSemaphore := make(chan struct{}, 30)

	chMsg := make(chan sdk.Msg)
	chErr := make(chan error)

	wg.Add(1)
	go func() {
		for i, seq := range seqs {
			wg.Add(1)
			chSemaphore <- struct{}{}

			go func(idx int, sq uint64) {
				defer func() {
					<-chSemaphore
					if idx == 0 {
						// to wait goroutine for closing channel
						wg.Done()
					}
					wg.Done()
				}()
				// concurrent logic
				p, err := senderChain.QueryPacket(senderQueryableHeight, sq)
				if err != nil {
					if strings.Contains(err.Error(), "Client.Timeout") {
						// TODO: skip
						log.Println("skip senderChain.QueryPacket() due to timeout: ", err)
						return
					}
					chErr <- err
					return
				}
				ack, err := receiverChain.QueryPacketAcknowledgement(receiverQueryableHeight, sq)
				if err != nil {
					if strings.Contains(err.Error(), "Client.Timeout") {
						// TODO: skip
						log.Println("skip receiverChain.QueryPacketAcknowledgement() due to timeout: ", err)
						return
					}
					chErr <- err
					return
				}
				res, err := receiverChain.QueryPacketAcknowledgementCommitmentWithProof(provableHeight, sq)
				if err != nil {
					chErr <- err
					return
				}

				chMsg <- chantypes.NewMsgAcknowledgement(*p, ack, res.Proof, res.ProofHeight, sender.String())
			}(i, seq)
		}
	}()
	go func() {
		wg.Wait()
		close(chMsg)
		close(chErr)
	}()

	// wait until channel is closed
	for {
		select {
		case msg, ok := <-chMsg:
			if !ok {
				return msgs, nil
			}
			msgs = append(msgs, msg)
		case err, ok := <-chErr:
			if ok {
				return nil, err
			}
		}
	}
}
