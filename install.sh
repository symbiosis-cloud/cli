#!/bin/bash
# shellcheck disable=SC2016
# based on https://raw.githubusercontent.com/canha/golang-tools-install-script/master/install.sh
set -e

VERSION="0.0.2"

OS="$(uname -s)"
ARCH="$(uname -m)"

case $OS in
    "Linux")
        case $ARCH in
        "x86_64")
            ARCH=amd64
            ;;
        "aarch64")
            ARCH=arm64
            ;;
        "armv8")
            ARCH=arm64
            ;;
        .*386.*)
            ARCH=386
            ;;
        esac
        PLATFORM="Linux_$ARCH"
    ;;
    "Darwin")
          case $ARCH in
          "x86_64")
              ARCH=amd64
              ;;
          "arm64")
              ARCH=arm64
              ;;
          esac
        PLATFORM="Darwin_$ARCH"
    ;;
esac

print_help() {
    echo "Usage: bash install.sh OPTIONS"
    echo -e "\nOPTIONS:"
    echo -e "  --remove\tRemove currently installed version"
    echo -e "  --version\tSpecify a version number to install"
}

if [ -z "$PLATFORM" ]; then
    echo "Your operating system is not supported by the script."
    exit 1
fi

if [ "$1" == "--remove" ]; then

    exit 0
elif [ "$1" == "--help" ]; then
    print_help
    exit 0
elif [ ! -z "$1" ]; then
    echo "Unrecognized option: $1"
    exit 1
fi

# if [ -d "$GOROOT" ]; then
#     echo "The Go install directory ($GOROOT) already exists. Exiting."
#     exit 1
# fi

BIN_LOCATION=$HOME/.symbiosis/bin

if [ ! -d "$BIN_LOCATION" ]; then
	mkdir -p "$BIN_LOCATION"
fi

TEMP_DIRECTORY=$(mktemp -d)
URL="https://github.com/symbiosis-cloud/cli/releases/download/v$VERSION/sym-cli_${VERSION}_$PLATFORM.tar.gz"

#https://github.com/symbiosis-cloud/cli/releases/download/v0.0.2/sym-cli_0.0.2_Darwin_arm64.tar.gz
#https://github.com/symbiosis-cloud/cli/releases/download/v0.0.2/sym-cli_Darwin_arm64.tar.gz

echo "Downloading $PACKAGE_NAME ..."
if hash wget 2>/dev/null; then
    wget -q $URL -O "$TEMP_DIRECTORY/sym.tar.gz"

else
    curl -so "$TEMP_DIRECTORY/sym.tar.gz" $URL
fi

if [ $? -ne 0 ]; then
    echo "Download failed! Exiting."
    exit 1
fi

echo "Extracting File..."

tar -C "$BIN_LOCATION" -xzf $TEMP_DIRECTORY/sym.tar.gz

exe="$BIN_LOCATION/sym"
rm -f "$TEMP_DIRECTORY/sym.tar.gz"

shell_profile=""
echo
echo "Symbiosis CLI was installed successfully to $exe"
if command -v sym >/dev/null; then
	echo "Run 'sym --help' to get started"
else

  if [ ! -w /usr/local/bin ]; then
    echo "Creating symlink in /usr/local/bin... asking for elevation"

    rc=0
    sudo ln -sf $exe /usr/local/bin/sym
  else
    ln -sf $exe /usr/local/bin/sym
  fi


  if ! command -v sym >/dev/null; then
    case $SHELL in
    /bin/zsh) shell_profile=".zshrc" ;;
    *) shell_profile=".bashrc" ;;
    esac

    echo
    echo "Symlink failed, probably due to access permissions. You can try running the installer again with sudo permissions."
    echo "Manually add the directory to your \$HOME/.bashrc (or similar)"
    echo "  export PATH=\"$BIN_LOCATION:\$PATH\""
  else
    echo "Run 'sym --help' to get started"
  fi

fi
echo

echo "Install completed"