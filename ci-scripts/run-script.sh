#!/usr/bin/env bash

if [[ "$TRAVIS_OS_NAME" != "windows" ]]; then
    make check-newcoin
    make lint
    make test-386
    make test-amd64
    make integration-tests-stable
    make lint-ui
    make build-ui-travis
    make test-ui
    make test-ui-e2e
fi

# if [[ "$TRAVIS_PULL_REQUEST" == false ]]; then
./ci-scripts/build-wallet.sh
# fi
