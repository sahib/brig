#!/bin/sh

pkill -9 brig
rm -rf /tmp/{ali,bob}
rm -f ~/.config/brig/registry.yml

# Have some color in the logs even when piping it somewhere else.
export BRIG_COLOR=always
alias brig-ali='brig --port 6666'
alias brig-bob='brig --port 6667'

# Uncomment the following lines to circumvent logging to syslog.
(brig-ali --repo /tmp/ali daemon launch -s 2>&1 > /tmp/log.ali) &
(brig-bob --repo /tmp/bob daemon launch -s 2>&1 > /tmp/log.bob) &

time brig-ali --repo /tmp/ali init ali -x > /dev/null
time brig-bob --repo /tmp/bob init bob -x > /dev/null

# Add them as remotes each
time brig-ali remote add bob $(brig-bob whoami -f)
time brig-bob remote add ali $(brig-ali whoami -f)

brig-ali stage BUGS ali-file
brig-ali commit -m 'Added ali-file'
brig-bob stage TODO bob-file
brig-bob commit -m 'Added bob-file'
