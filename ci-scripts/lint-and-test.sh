#!/usr/bin/env bash

set -e -o pipefail

make check-newcoin
make lint

if [[ ${TEST_SUIT} == "units" ]]; then
    make test-386
    make test-amd64
elif [[ ${TEST_SUIT} == "integration" ]]; then
    make integration-tests-stable
fi

make lint-ui
make build-ui-travis
make test-ui
make test-ui-e2e
