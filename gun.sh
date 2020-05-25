#!/bin/bash

SRC_CHAIN_FILE="configs/demo/ibc0.json"
DST_CHAIN_FILE="configs/demo/ibc1.json"
AMOUNT="1n0token"

port_id="transfer"

src_chain_id=$(jq -r '."chain-id"' $SRC_CHAIN_FILE)
dst_chain_id=$(jq -r '."chain-id"' $DST_CHAIN_FILE)

rgun cfg init

rgun ch a -f $SRC_CHAIN_FILE
rgun ch a -f $DST_CHAIN_FILE

rgun keys restore ibc0 testkey "$(jq -r '.secret' data/ibc0/n0/gaiacli/key_seed.json)"
rgun keys restore ibc1 testkey "$(jq -r '.secret' data/ibc1/n0/gaiacli/key_seed.json)"

rgun lite init $src_chain_id -f
rgun lite init $dst_chain_id -f

rgun config show

rgun paths gen $src_chain_id $port_id $dst_chain_id $port_id "$src_chain_id-$dst_chain_id"

rgun tx link "$src_chain_id-$dst_chain_id"

#rgun tx gun $src_chain_id $dst_chain_id $AMOUNT true $(rgun ch addr $dst_chain_id) 10 --gas 1000000 -d
#rgun tx gun $dst_chain_id $src_chain_id $AMOUNT false $(rgun ch addr $src_chain_id) 10 --gas 1000000 -d
#rgun tx gun_update_client $src_chain_id $dst_chain_id 5s