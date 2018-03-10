#!/usr/bin/env bash
set -e

echo Unit Tests
echo

go test ./...

echo
echo Integration Test
echo
TMP_DIR=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`
PLUGIN_BIN_PATH=${TMP_DIR}/plugin

# Compile the plugin
go build -o ${PLUGIN_BIN_PATH} cmd/plugin.go

# Install it

CF_HOME=${TMP_DIR} cf install-plugin ${PLUGIN_BIN_PATH} -f
