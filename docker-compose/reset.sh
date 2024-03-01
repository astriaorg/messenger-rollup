#!/bin/bash

CURRENT_DIR=$(dirname "$0")

rm -rf $CURRENT_DIR/.data
mkdir -p $CURRENT_DIR/.data/cometbft
mkdir -p $CURRENT_DIR/.data/sequencer

# Reset the .data/cometbft/priv_validator_state.json file
echo '{
  "height": "0",
  "round": 0,
  "step": 0
}' > $CURRENT_DIR/.data/cometbft/priv_validator_state.json
