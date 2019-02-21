#!/bin/sh

USE_SINGLE=false

while getopts ":s" opt; do
  case $opt in
    s)
      USE_SINGLE=true
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
  esac
done

# Kill previous test-bed
pkill -9 brig
rm -rf /tmp/{ali,bob}

# Have some color in the logs even when piping it somewhere else.
export BRIG_COLOR=always
alias brig-ali='brig --port 6666 --repo /tmp/ali'

if [ "$USE_SINGLE" = false ]; then
    alias brig-bob='brig --port 6667 --repo /tmp/bob'
else
    # Fake the invocation of bob's brig.
    alias brig-bob='/bin/true'
fi

# Give the daemon to start up a bit.
(brig-ali init ali -x > /dev/null) &
(brig-bob init bob -x > /dev/null) &

sleep 1

# Add them as remotes each
if [ "$USE_SINGLE" = false ]; then
    brig-ali remote add bob $(brig-bob whoami -f)
    brig-bob remote add ali $(brig-ali whoami -f)
fi

brig-ali stage BUGS ali-file
brig-ali commit -m 'added ali-file'
brig-bob stage LICENSE bob-file
brig-bob commit -m 'added bob-file'
