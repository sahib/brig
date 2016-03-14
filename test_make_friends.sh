#!/bin/sh

echo -n "alice@jabber.nullcat.de/laptop: " > /tmp/bob/.brig/otr.buddies
BRIG_PATH=/tmp/alice brig auth -p >> /tmp/bob/.brig/otr.buddies

echo -n "bob@jabber.nullcat.de/desktop: " > /tmp/alice/.brig/otr.buddies
BRIG_PATH=/tmp/bob brig auth -p >> /tmp/alice/.brig/otr.buddies
