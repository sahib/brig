#!/bin/sh

BRIG_PATH=/tmp/alice brig auth -a bob@jabber.nullcat.de/desktop $(BRIG_PATH=/tmp/bob brig auth -p)
BRIG_PATH=/tmp/bob brig auth -a alice@jabber.nullcat.de/laptop $(BRIG_PATH=/tmp/alice brig auth -p)
