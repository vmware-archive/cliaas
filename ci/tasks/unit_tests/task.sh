#!/bin/bash -ue
export GOPATH=/go
export PATH=$GOPATH/bin:$PATH
go get github.com/Masterminds/glide
mkdir -p /go/src/github.com/c0-ops
cp -r ./cliaas/ /go/src/github.com/pivotal-cf/cliaas
cd /go/src/github.com/pivotal-cf/cliaas
glide install
ignoreVendorAndCI=$(glide nv | grep -v './integration_tests/...')
go test ${ignoreVendorAndCI} -v
