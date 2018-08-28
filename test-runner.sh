#!/bin/bash

# Run the test cases (name test-xyz.sh in tests/)
# inside of a docker container. Each of them is
# completely separated and does no harm to the host.

TEST_FAIL_OUTPUT=/tmp/brig-docker.log

run_test() {
    echo "-- EXECUTE: " $1
    docker run -it -v $(pwd):/tmp/tests brig bash /tmp/tests/$1 2>&1 > ${TEST_FAIL_OUTPUT}
    if [[ 0 -ne $? ]]; then
        echo "** FAILED: " $1 "exited with non-zero code. Container output:"
        cat ${TEST_FAIL_OUTPUT}
        echo "==========="
    else
        echo "** SUCCESS:" $1
        echo "==========="
    fi
}

for arg in "$@"
do
    if [[ ${arg} = "help" ]]; then
        echo "usage: test-runner.sh [test_name...]"
        echo "       when specifying no argument, all tests are run."
        exit 0
    fi
done

cd tests

if [[ $# -eq 0 ]]; then
    for path in $(ls test-* ); do
        run_test ${path}
    done
else
    for name in "$@"
    do
        run_test ${name}
    done
fi
