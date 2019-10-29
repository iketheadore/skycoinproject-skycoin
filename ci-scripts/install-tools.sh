set -e -o pipefail


if [[ "$TRAVIS_OS_NAME" == "windows" ]]; then
    # Install nodejs with choco
    choco install nodejs --version=8.11.0 -y
    echo 'export PATH="/c/Program Files/nodejs:${PATH}";' >> ~/.bashrc
else
    #. $HOME/.nvm/nvm.sh  # This loads NVM
    nvm install 8.11.0
    nvm use 8.11.0
    make install-linters
    make install-deps-ui
fi
