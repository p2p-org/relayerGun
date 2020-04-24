rly config init

# NOTE: you may want to look at the config between these steps to see
# what is added in each step. The config is located at ~/.relayer/config/config.yaml
cat ~/.relayer/config/config.yaml

# Then add the chains and paths that you will need to work with the
# gaia chains spun up by the two-chains script
rly cfg add-dir configs/demo/

# NOTE: you may want to look at the config between these steps
cat ~/.relayer/config/config.yaml

# Now, add the key seeds from each chain to the relayer to give it funds to work with
rly keys restore ibc0 testkey "$(jq -r '.secret' data/ibc0/n0/gaiacli/key_seed.json)"
rly keys restore ibc1 testkey "$(jq -r '.secret' data/ibc1/n0/gaiacli/key_seed.json)"

# Then its time to initialize the relayer's lite clients for each chain
# All data moving forward is validated by these lite clients.
rly lite init ibc0 -f
rly lite init ibc1 -f

# At this point the relayer --home directory is ready for normal operations between
# ibc0 and ibc1. Looking at the folder structure of the relayer at this point is helpful
tree ~/.relayer

# Now you can connect the two chains with one command:
rly tx link demo

# Check the token balances on both chains
rly q balance ibc0
rly q bal ibc1

# Then send some tokens between the chains
rly tx transfer ibc0 ibc1 10000n0token true $(rly keys show ibc1 testkey)

# See that the transfer has completed
rly q bal ibc0
rly q bal ibc1

# Send the tokens back to the account on ibc0
rly tx xfer ibc1 ibc0 10000n0token false $(rly keys show ibc0 testkey)

# See that the return trip has completed
rly q bal ibc0
rly q bal ibc1

# NOTE: you will see the stake balances decreasing on each chain. This is to pay for fees
# You can change the amount of fees you are paying on each chain in the configuration.