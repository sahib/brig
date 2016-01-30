#!/bin/bash
export BRIG_PATH=/tmp/alice
echo "Im a dummy file full of fun." > hello_world 

function hash() {
    sha1sum | base64
}

function section() {
    sleep 0.1 && echo $1
}

section "=== WAIT ==="
brig daemon-wait

section "=== ADD ==="
brig add hello_world
section "=== CAT ==="
brig cat hello_world | tee >(hash)
section "=== MOUNT ==="
mkdir -p /tmp/mount
brig mount /tmp/mount
cat /tmp/mount/hello_world | tee >(hash)
# section "=== MODIFY ==="
# echo 'Even more fun.' >> hello_world
# brig add hello_world
# brig cat hello_world | tee >(hash)
# cat /tmp/mount/hello_world | tee >(hash)
section "=== FUSE MODIFY ==="
echo 'What now?' > /tmp/mount/hello_world
cat /tmp/mount/hello_world | tee >(hash)
section "=== FINISH ==="
