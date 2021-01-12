#!/bin/bash

set -e

export GOMAXPROCS=20

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
pkill -9 brig || true
rm -rf /tmp/{ali,bob}
rm -rf /tmp/brig.socket*

# Have some color in the logs even when piping it somewhere else.
export BRIG_COLOR=always
brig_ali() {
    brig --repo /tmp/ali "$@"
}

brig_bob() {
    if [ "$USE_SINGLE" = false ]; then
        brig --repo /tmp/bob "$@"
    fi
}

brig_ali init ali --ipfs-path /tmp/ali-ipfs
brig_bob init bob --ipfs-path /tmp/bob-ipfs

# Add them as remotes each
if [ "$USE_SINGLE" = false ]; then
    # shellcheck disable=SC2046
    brig_ali remote add bob $(brig_bob whoami -f)
    # shellcheck disable=SC2046
    brig_bob remote add ali $(brig_ali whoami -f)
fi

brig_ali -V stage TODO ali-file
brig_ali commit -m 'added ali-file'
brig_bob stage LICENSE bob-file
brig_bob commit -m 'added bob-file'
