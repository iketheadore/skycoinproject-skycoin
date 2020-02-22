#!/usr/bin/env bash

set -e -o pipefail

make check-newcoin
make lint

if [[ ${TEST_SUIT} == "units" ]]; then
    echo "DO unit tests"
    make test-386
    make test-amd64
elif [[ ${TEST_SUIT} == "integrations" ]]; then
    echo "DO integration tests"
    make integration-tests-stable
fi

make lint-ui
make build-ui-travis
make test-ui
make test-ui-e2e
