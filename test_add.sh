#!/bin/sh

export BRIG_PATH=/tmp/alice
pkill -f brig
rm -rf $BRIG_PATH
echo "=== INIT ==="
brig init alice@jabber.de/home -x ThiuJ9wesh --nodaemon
echo "=== DAEMON ==="
brig daemon -x ThiuJ9wesh
echo "=== FINISH ==="
