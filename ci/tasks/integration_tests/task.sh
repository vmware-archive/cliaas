#!/bin/bash -ue

cd /go/src/github.com/pivotal-cf/cliaas
glide install
go test ./integration_tests/$IAAS/... -v
