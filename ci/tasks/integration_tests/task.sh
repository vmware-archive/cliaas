#!/bin/bash -ue

pushd /go/src/github.com/c0-ops/cliaas
glide install
go test ./integration_tests/$IAAS/... -v
