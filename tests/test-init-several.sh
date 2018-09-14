#!/bin/bash

# unset BRIG_PATH, since it's confusing for several.
export BRIG_PATH=

brig --repo "/tmp/ali" --verbose init ali -w "echo ali"
brig --repo "/tmp/bob" --verbose init bob -w "echo bob"
brig --repo "/tmp/cem" --verbose init cem -w "echo cem"

FILE_COUNT=$(ls /tmp/ali | wc -l)
if [[ $FILE_COUNT -eq 0 ]]; then
    echo "!! /tmp/ali empty after init"
    exit 1
fi

FILE_COUNT=$(ls /tmp/bob | wc -l)
if [[ $FILE_COUNT -eq 0 ]]; then
    echo "!! /tmp/bob empty after init"
    exit 2
fi

FILE_COUNT=$(ls /tmp/cem | wc -l)
if [[ $FILE_COUNT -eq 0 ]]; then
    echo "!! /tmp/cem empty after init"
    exit 3
fi
