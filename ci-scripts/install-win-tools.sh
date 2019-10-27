# Install nodejs with choco
choco install nodejs --version=8.11.0 -y
touch ~/.bashrc
echo "export PATH='$PATH:/c/Program Files/nodejs/'" >> ~/.bashrc
