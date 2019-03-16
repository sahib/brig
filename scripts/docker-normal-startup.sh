#!/bin/bash

export BRIG_USER="bob@wonderland.lit/container"
brig --verbose init $BRIG_USER -x || exit 1
brig --verbose -x --bind 0.0.0.0 daemon launch --log-to-stdout
