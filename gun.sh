#!/bin/bash

SRC_CHAIN_FILE="wowlol.json"
DST_CHAIN_FILE="lolchain.json"
AMOUNT="10sometoken"

port_id="transfer"

src_chain_id=$(jq -r '."chain-id"' $SRC_CHAIN_FILE)
dst_chain_id=$(jq -r '."chain-id"' $DST_CHAIN_FILE)

N=2

home=home/.relayer

for ((i=0; i<$N; i++))
do
  rly --home $home$i cfg init

  rly --home $home$i ch a -f $SRC_CHAIN_FILE
  rly --home $home$i ch a -f $DST_CHAIN_FILE

  rly --home $home$i lite init $src_chain_id -f
  rly --home $home$i lite init $dst_chain_id -f

  rly --home $home$i chains list

  rly --home $home$i keys add $src_chain_id
  rly --home $home$i keys add $dst_chain_id

  rly --home $home$i tst req $src_chain_id
  rly --home $home$i tst req $dst_chain_id

  rly --home $home$i paths gen $src_chain_id $port_id $dst_chain_id $port_id "$src_chain_id-$dst_chain_id"

  rly --home $home$i tx link "$src_chain_id-$dst_chain_id"

  echo "Starting gun..."
  rly --home $home$i tx gun $src_chain_id $dst_chain_id $AMOUNT true $(rly --home $home$i ch addr $dst_chain_id) > $home$i.log 2>&1 &
  echo "Started!"

done

#for (( ; ; ))
#do
#    rly --home $home$i tx transfer $src_chain_id $dst_chain_id $AMOUNT true $(rly --home $home$i ch addr $dst_chain_id)
#    rly --home $home$i q bal $src_chain_id
#    rly --home $home$i q bal $dst_chain_id
#done