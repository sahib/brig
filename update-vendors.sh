#!/bin/bash

echo "1) Updating non-ipfs deps..."
dep ensure -v

echo "2) Update ipfs to the latest version..."
go get -u github.com/ipfs/go-ipfs 
cd $GOPATH/src/github.com/ipfs/go-ipfs
make install

echo "3) Take updated ipfs and commit to vendor..."
cd $GOPATH/github.com/disorganizer/brig/vendor

# This is an insanely hacky way to do this, but it kinda works out.
# Also, this would not be necessary if ipfs' would get rid of gx
# and use a sane vendor package as the rest of the world does.
rm github.com/ipfs/go-ipfs gx .git -rf
cp -r $GOPATH/src/github.com/ipfs/go-ipfs github.com/ipfs
cp -r $GOPATH/src/gx github.com/ipfs .
git add .
git commit -am 'updated vendor repository'
git remote add origin git@github.com:disorganizer/brig-vendor.git

echo "4) Uploading vendor repository..."
push --set-upstream origin master --force

echo "5) Update vendor submodule in main repo..."
cd ..
git add vendor
git commit -m "updated vendor/"
