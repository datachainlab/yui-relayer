ARG GETH_VERSION

FROM ethereum/client-go:${GETH_VERSION}

ARG CHAINID
ARG NETWORKID

ADD geth.password /root/geth.password
ADD genesis.json  /root/genesis.json
ADD geth-config.toml /root/geth-config.toml
ADD privatekey  /root/privatekey

RUN sed -i "s#\"chainId\": 0#\"chainId\": $CHAINID#g" /root/genesis.json
RUN sed -i "s#NetworkId = 0#NetworkId = $NETWORKID#g" /root/geth-config.toml

RUN mkdir -p /root/backup
RUN /usr/local/bin/geth --datadir /root/.ethereum init /root/genesis.json
RUN /usr/local/bin/geth account import /root/privatekey --password /root/geth.password

EXPOSE 8545 8546 30303

ENTRYPOINT ["geth"]
CMD ["--config", "/root/geth-config.toml", "--mine", "--miner.threads", "1"]
