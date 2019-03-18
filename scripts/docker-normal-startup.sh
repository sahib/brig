#!/bin/bash

set -e

brig --verbose init -x $BRIG_USER
brig daemon quit
sleep 2
brig daemon launch -s
