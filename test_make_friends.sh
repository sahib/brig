#!/bin/sh

ALICE_ID=$(BRIG_PATH=/tmp/alice brig remote self | cut -d " " -f 1)
BOB_ID=$(BRIG_PATH=/tmp/bob brig remote self | cut -d " " -f 1)

BRIG_PATH=/tmp/alice brig remote add bob@jabber.nullcat.de/desktop $BOB_ID
BRIG_PATH=/tmp/bob brig remote add alice@jabber.nullcat.de/laptop $ALICE_ID
