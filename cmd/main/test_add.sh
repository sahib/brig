#!/bin/sh

export BRIG_PATH=/tmp/alice
pkill -f brig
rm -rf $BRIG_PATH
echo "=== INIT ==="
./brig init alice@jabber.de/home -x hello_password --nodaemon
echo "=== DAEMON ==="
./brig daemon -x hello_password
echo "=== FINISH ==="
