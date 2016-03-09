#!/bin/sh

export BRIG_PATH=/tmp/alice
rm -rf $BRIG_PATH
echo "=== INIT ==="
brig init alice@jabber.nullcat.de/laptop -x ThiuJ9wesh --nodaemon
echo "=== DAEMON ==="
brig daemon -x ThiuJ9wesh
echo "=== FINISH ==="
