#!/bin/bash

# Standard values:
BRIG_PORT=6669
BRIG_USER=bob
BRIG_PATH=/tmp/test-$BRIG_USER-repo

BACKEND=ipfs

while getopts ":u:p:a:b" opt; do
  case $opt in
    u)
      BRIG_USER=$OPTARG
      BRIG_PATH=/tmp/test-$BRIG_USER-repo
      ;;
    p)
      BRIG_PORT=$OPTARG
      ;;
    a)
      BRIG_PATH=$OPTARG
      ;;
    b)
       BACKEND=$OPTARG
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
    :)
      echo "Option -$OPTARG requires an argument." >&2
      exit 1
      ;;
  esac
done

echo "== Initializing brig instance..."
rm -rf $BRIG_PATH
brig -x init $BRIG_USER --backend mock

PARENT_SHELL=$(ps -o comm= $PPID | cut -d' ' -f 2)
printf "== Launching shell '%s' with BRIG_PATH=%s and BRIG_PORT=%d\n" \
    "$PARENT_SHELL" \
    "$BRIG_PATH" \
    "$BRIG_PORT"

printf "== Exit the shell once you are done to cleanup.\n"

# Let the user use the shell:
# (In russia child process starts YOU!)
exec $PARENT_SHELL

# Try to clean up by exiting the daemon:
echo "== Cleaning up..."
brig -x daemon quit
