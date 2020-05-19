FROM golang:1.14-alpine

RUN apk add --update --no-cache bash ca-certificates git libc-dev make build-base jq
ENV RELAYERPATH /go/src/github.com/p2p-org/relayerGun
WORKDIR $RELAYERPATH

COPY . .

ENV GO111MODULE=on
RUN make install

ENTRYPOINT $RELAYERPATH/relay.sh