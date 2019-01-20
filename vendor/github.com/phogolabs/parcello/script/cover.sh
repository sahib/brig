#!/bin/bash

echo "mode: atomic" > coverage.txt

for profile in $(find . -name '*.coverprofile' -maxdepth 10  -type f); do
  grep -v "mode: " < "$profile" >> coverage.txt
  rm "$profile"
done
