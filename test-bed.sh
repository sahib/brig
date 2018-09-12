#!/bin/sh

pkill -9 brig
rm -rf /tmp/{ali,bob}
rm ~/.config/brig/registry.yml

alias brig-ali='brig -p 6666'
alias brig-bob='brig -p 6667'

brig-ali --repo /tmp/ali init ali -x > /dev/null
brig-bob --repo /tmp/bob init bob -x > /dev/null

brig-ali remote add bob $(brig-bob whoami -f)
brig-bob remote add ali $(brig-ali whoami -f)

brig-ali stage BUGS
brig-bob stage TODO
