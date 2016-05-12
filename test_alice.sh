#!/bin/sh

export BRIG_PATH=/tmp/alice
rm -rf $BRIG_PATH
echo "=== INIT ==="
brig -x ThiuJ9wesh --nodaemon init alice@jabber.nullcat.de/laptop 
echo "=== DAEMON ==="
brig -x ThiuJ9wesh daemon launch
echo "=== FINISH ==="
