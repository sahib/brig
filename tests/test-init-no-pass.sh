#!/bin/bash

mkdir -p $BRIG_PATH
cd $BRIG_PATH

FILE_COUNT=$(ls $BRIG_PATH | wc -l)
if [[ $FILE_COUNT -ne 0 ]]; then
    echo "!! /tmp/repo is not empty"
    exit 1
fi

# Use no password
brig -x --verbose init alice

FILE_COUNT=$(ls $BRIG_PATH | wc -l)
if [[ $FILE_COUNT -eq 0 ]]; then
    echo "!! /tmp/repo empty after init"
    exit 1
fi
