# !/usr/bin/env bash

set -x -e -o pipefail

# Install nodejs with choco
# choco install nodejs --version=8.11.0 -y
touch ~/.bashrc
echo "export PATH='/c/Program Files/nodejs/:\$PATH'" >> ~/.bashrc

echo $PATH
