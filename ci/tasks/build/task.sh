#!/bin/bash -ue
export GOPATH=/go
export PATH=$GOPATH/bin:$PATH
go get github.com/Masterminds/glide
mkdir -p /go/src/github.com/c0-ops
cp -r ./cliaas/ /go/src/github.com/c0-ops/cliaas
cd /go/src/github.com/c0-ops/cliaas
glide install

version=beta
output_path=cliaas-linux
go build \
    -o $output_path \
    -ldflags "-s -w -X cliaas.Version=${version}" \
    cmd/cliaas/main.go
