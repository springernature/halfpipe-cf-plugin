#!/usr/bin/env sh
set -e


if [ ! -z "${RUNNING_IN_CI}" ]; then
    echo "Installing and putting CF on path"
    TMPDIR=`mktemp -d`
    CF_TAR_URL="https://packages.cloudfoundry.org/stable?release=linux64-binary&version=6.35.0&source=github-rel"
    wget -qO- ${CF_TAR_URL} | tar xvz -C $TMPDIR
    export PATH=$PATH:$TMPDIR

    echo "Overiding GOPATH to current dir"
    export GOPATH=`pwd`
fi

echo Unit Tests
echo

go test ./...

echo
echo Integration Test
echo

TMP_DIR=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'` # Support both for linux and osx..
PLUGIN_BIN_PATH=${TMP_DIR}/plugin

# Compile the plugin
go build -o ${PLUGIN_BIN_PATH} cmd/plugin.go

# Install it

CF_HOME=${TMP_DIR} cf install-plugin ${PLUGIN_BIN_PATH} -f
