#!/bin/sh

export BRIG_PATH=/tmp/bob
rm -rf $BRIG_PATH
echo "=== INIT ==="
brig -x eecot3oXan --nodaemon init bob@jabber.nullcat.de/desktop 
echo "=== DAEMON ==="
brig -x eecot3oXan daemon launch
echo "=== FINISH ==="
