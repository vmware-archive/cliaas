#!/bin/bash -ue
export GOPATH=$PWD/go
export PATH=$GOPATH/bin:$PATH
cd go/src/github.com/c0-ops/cliaas
go get github.com/Masterminds/glide
glide install
go test $(glide nv) -v
