#!/bin/bash

DATA_DIR=${DATA_DIR:-"data"}
CHAIN_ID0=${CHAIN_ID:-"ibc0"}
CHAIN_ID1=${CHAIN_ID:-"ibc1"}

rlyhome_demo="$DATA_DIR/relayer/demo"
rlyhome_demoNFT="$DATA_DIR/relayer/demoNFT"

while test $# -gt 0; do
  case "$1" in
    -h|--help)
      echo "run relayer"
      echo " "
      echo "relay.sh [options]"
      echo " "
      echo "options:"
      echo "--init                   force reset blockchain state (clear all data)"
      echo "--norun                   omit service start"
      exit 0
      ;;
    --init)
      INIT=true
      shift
      ;;
    --norun)
      NORUN=true
      shift
      ;;
    *)
      break
      ;;
  esac
done

if [[ $INIT ]]; then
  echo "Clearing data..."
  rm -rf $rlyhome &> /dev/null
fi


if [ ! -d "$rlyhome" ]; then
  rly --home $rlyhome_demo config init
  rly --home $rlyhome_demoNFT config init

# Then add the chains and paths that you will need to work with the
# gaia chains spun up by the two-chains script
rly --home $rlyhome_demo cfg add-dir demoIBC/
rly --home $rlyhome_demoNFT cfg add-dir demoIBC/

# NOTE: you may want to look at the config between these steps
#cat ~/.relayer/config/config.yaml

# Now, add the key seeds from each chain to the relayer to give it funds to work with
#rly keys restore ibc0 testkey "$(head -n 1 data_r0.txt)"
#rly keys restore ibc1 testkey "$(head -n 1 data_r1.txt)"

rly --home $rlyhome_demo keys restore $CHAIN_ID0 testkey "$(jq -r '.secret' "$DATA_DIR/$CHAIN_ID0/n0/gaiacli/key_seed.json")"
rly --home $rlyhome_demo keys restore $CHAIN_ID1 testkey "$(jq -r '.secret' "$DATA_DIR/$CHAIN_ID1/n0/gaiacli/key_seed.json")"
rly --home $rlyhome_demoNFT keys restore $CHAIN_ID0 testkey "$(jq -r '.secret' "$DATA_DIR/$CHAIN_ID0/n0/gaiacli/key_seed.json")"
rly --home $rlyhome_demoNFT keys restore $CHAIN_ID1 testkey "$(jq -r '.secret' "$DATA_DIR/$CHAIN_ID1/n0/gaiacli/key_seed.json")"

# Then its time to initialize the relayer's lite clients for each chain
# All data moving forward is validated by these lite clients.
rly --home $rlyhome_demo lite init ibc0 -f
rly --home $rlyhome_demo lite init ibc1 -f
rly --home $rlyhome_demoNFT lite init ibc0 -f
rly --home $rlyhome_demoNFT lite init ibc1 -f

rly --home $rlyhome_demo tx link demo
rly --home $rlyhome_demoNFT tx link demoNFT

fi

if [[ -z $NORUN ]]; then
  echo "Starting relayer..."

  rly --home $rlyhome_demo start demo & #> rly_demo.log 2>&1 &
  rly --home $rlyhome_demoNFT start demoNFT #> rly_demoNFT.log 2>&1 &
fi

