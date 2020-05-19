#!/bin/bash
if [ ! -d ~/data/relayer ]; then
  mkdir -p ~/data/relayer/demo
  mkdir -p ~/data/relayer/demoNFT
fi

docker build -t relayer .
docker network create dwh-network
docker-compose -f demo.yml up -d
