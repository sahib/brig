#!/bin/bash

mkdir -p $BRIG_PATH
cd $BRIG_PATH

FILE_COUNT=$(ls $BRIG_PATH | wc -l)
if [[ $FILE_COUNT -ne 0 ]]; then
    echo "!! $BRIG_PATH is not empty"
    exit 1
fi

# Use a password helper:
brig --verbose init alice -w "echo mypass"

FILE_COUNT=$(ls $BRIG_PATH | wc -l)
if [[ $FILE_COUNT -eq 0 ]]; then
    echo "!! $BRIG_PATH empty after init"
    exit 2
fi

# Check if we can really access stuff:
brig cat README.md | grep $BRIG_PATH
if [[ $? -ne 0 ]]; then
    echo "!! No readme.md was created"
    exit 3
fi

# Check that we can also restart with a problem:
brig daemon quit
sleep 2
brig ls
