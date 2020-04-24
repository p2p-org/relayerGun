#!/bin/bash

SRC_CHAIN_FILE="wowlol.json"
DST_CHAIN_FILE="lolchain.json"
AMOUNT="10sometoken"

port_id="transfer"

src_chain_id=$(jq -r '."chain-id"' $SRC_CHAIN_FILE)
dst_chain_id=$(jq -r '."chain-id"' $DST_CHAIN_FILE)

N=5

#rly cfg init
#
#rly ch a -f $SRC_CHAIN_FILE
#rly ch a -f $DST_CHAIN_FILE
#
#rly lite init $src_chain_id -f
#rly lite init $dst_chain_id -f
#
#rly keys add $src_chain_id
#rly tst req $src_chain_id
#
#rly keys add $dst_chain_id
#rly tst req $dst_chain_id
#
#for ((i=0; i<$N; i++))
#do
#
#  rly keys add $src_chain_id "key$i"
#  rly tst req $src_chain_id "key$i"
#
#done
#
#rly config show
#
#rly paths gen $src_chain_id $port_id $dst_chain_id $port_id "$src_chain_id-$dst_chain_id"
#
#rly tx link "$src_chain_id-$dst_chain_id"

rly tx gun $src_chain_id $dst_chain_id $AMOUNT true $(rly ch addr $dst_chain_id) -d