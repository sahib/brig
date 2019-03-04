#!/bin/bash

export BRIG_USER="bob@wonderland.lit/container"
brig --verbose -x init $BRIG_USER || exit 1
brig --verbose -x --bind 0.0.0.0 daemon launch --log-to-stdout
