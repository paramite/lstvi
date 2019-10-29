#!/bin/bash
set -ex

# bootstrap
mkdir -p /go/bin /go/src /go/pkg
export GOPATH=/go
export PATH=$PATH:$GOPATH/bin

# get dependencies
yum install -y epel-release
yum install -y golang
go get -u github.com/golang/dep/...
dep ensure -v --vendor-only

# run unit test suite
go test -v ./tests/

# run benchmark
go test -bench=. ./tests/
