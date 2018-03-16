#!/usr/bin/env sh
set -e

# If you want to run this before pushing
# mkdir -p /opt/resource
# chown -R yourUser:yourUser /opt/resource
# go build -o /opt/resource/out cmd/out/out.go
# export API=google-api
# export USERNAME=engineering-enablement user
# export PASSWORD=asdasd
# export ORG=engineering-enablement
# export SPACE=integration_test

go run .integration_test/integration.go `pwd`

