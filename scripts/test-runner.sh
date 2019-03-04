#!/bin/bash

# Run the test cases (name test-xyz.sh in tests/)
# inside of a docker container. Each of them is
# completely separated and does no harm to the host.

TEST_FAIL_OUTPUT=/tmp/brig-docker.log
FOLLOW_OUTPUT=false

run_test() {
    echo "-- EXECUTE: " $1

	if [[ ${FOLLOW_OUTPUT} = true ]]; then
    	docker run -it -v $(pwd):/tmp/tests brig bash /tmp/tests/$1
		TEST_FAIL_OUTPUT="// see above //"
	else
    	docker run -it -v $(pwd):/tmp/tests brig bash /tmp/tests/$1 2>&1 > ${TEST_FAIL_OUTPUT}
	fi

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
    if [[ ${arg} = "--help" ]]; then
        echo "usage: test-runner.sh [test_name...]"
        echo "       when specifying no argument, all tests are run."
        exit 0
    fi
    if [[ ${arg} = "-v" ]] || [[ ${arg} = "--verbose" ]]; then
		FOLLOW_OUTPUT=true
	fi
done

cd tests

if [[ $# -eq 0 ]]; then
	echo "-- Running all tests."
    for path in $(ls test-* ); do
        run_test ${path}
    done
else
    for name in "$@"
    do
		# Filter arguments:
		if [[ ${name} != \-* ]]; then
        	run_test ${name}
		fi
    done
fi
