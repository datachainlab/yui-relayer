ARG GETH_VERSION

FROM ethereum/client-go:${GETH_VERSION}

WORKDIR /root

ARG NETWORKID
ARG LEDGER_BACKUP_PATH

COPY ${LEDGER_BACKUP_PATH}/backup.tar ./backup.tar
RUN tar -xf backup.tar -C /root
RUN rm -rf backup.tar

COPY contract/build/contracts /root/contracts
COPY Dockerfiles/geth-scaffold/geth-config.toml /root/geth-config.toml
COPY Dockerfiles/geth-scaffold/geth.password /root/geth.password
COPY Dockerfiles/geth-scaffold/mnemonic /root/mnemonic
COPY Dockerfiles/geth-scaffold/privatekey /root/privatekey

RUN sed -i "s#NetworkId = 0#NetworkId = $NETWORKID#g" /root/geth-config.toml

EXPOSE 8545 8546 30303

ENTRYPOINT ["geth"]
CMD ["--config", "/root/geth-config.toml", "--mine", "--miner.threads", "1"]
