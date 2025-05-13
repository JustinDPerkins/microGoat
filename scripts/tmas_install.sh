#!/bin/bash

# Check if JQ is installed.
if ! command -v jq &> /dev/null
then
    echo "JQ could not be found."
    exit
fi

# Check if curl is installed.
if ! command -v curl &> /dev/null
then
    echo "curl could not be found."
    exit
fi

# Check if sudo is installed.
if ! command -v sudo &> /dev/null
then
    echo "sudo could not be found."
    exit
fi

BASE_URL="https://cli.artifactscan.cloudone.trendmicro.com/tmas-cli/"
METADATA_URL="${BASE_URL}metadata.json"
VERSION_STRING=$(curl -s $METADATA_URL | jq -r '.latestVersion')
VERSION="${VERSION_STRING:1}"
echo "Latest version is: $VERSION"

OS=$(uname -s)
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ]; then ARCH=arm64; fi
ARCHITECTURE="${OS}_${ARCH}"
echo "Downloading version $VERSION of tmas CLI for $OS in architecture $ARCHITECTURE"
if [ "$OS" = "Linux" ]; 
then 
    URL="${BASE_URL}latest/tmas-cli_$ARCHITECTURE.tar.gz"
    curl -s "$URL" | tar -xz tmas
else
    URL="${BASE_URL}latest/tmas-cli_$ARCHITECTURE.zip"
    curl -s "$URL" -o tmas.zip
    unzip -p tmas.zip tmas > extracted_tmas
    mv extracted_tmas tmas
    chmod +x tmas
    rm -rf tmas.zip
fi

echo "Moving the binary to \"/usr/local/bin/\". It might request root access."
sudo mv tmas /usr/local/bin/

# If v1cs is already installed, create a symbolic link to tmas to maintain compatibility.
if command -v c1cs &> /dev/null
then
    echo "Creating symbolic link from c1cs to tmas to maintain compatibility. Note: this might be removed in the future."
    sudo ln -sf /usr/local/bin/tmas /usr/local/bin/c1cs
fi
