#!/bin/bash

# Simulate installed packages
INSTALLED_PACKAGES=""

# Function to check if a package is "installed"
is_installed() {
    if [[ " $INSTALLED_PACKAGES " == *" $1 "* ]]; then
        return 0
    else
        return 1
    fi
}

case "$1" in
    "list-installed")
        # Extract package name from grep pattern
        PKG_NAME=$(echo "$2" | sed -n "s/^'^\(.*\)\s'$/\1/p")
        if is_installed "$PKG_NAME"; then
            echo "$PKG_NAME - mock-version"
            exit 0
        else
            exit 1 # Not installed
        fi
        ;;
    "update")
        echo "Mock opkg update successful"
        exit 0
        ;;
    "install")
        PKG_TO_INSTALL="$2"
        echo "Mock opkg install $PKG_TO_INSTALL successful"
        INSTALLED_PACKAGES+=" $PKG_TO_INSTALL" # Simulate installation
        exit 0
        ;;
    *)
        echo "Mock opkg: Unknown command $*" >&2
        exit 1
        ;;
esac