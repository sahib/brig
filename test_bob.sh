#!/bin/sh

export BRIG_PATH=/tmp/bob
rm -rf $BRIG_PATH
echo "=== INIT ==="
brig init bob@jabber.nullcat.de/desktop -x eecot3oXan --nodaemon
echo "=== DAEMON ==="
brig daemon -x eecot3oXan
echo "=== FINISH ==="
