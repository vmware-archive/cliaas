#!/usr/bin/env bash

set -ex

export GOPATH=$PWD/go

root=$PWD
output_path=$root/$OUTPUT_PATH
version=$(cat $root/version/version)

cd go/src/github.com/c0-ops/cliaas
  go build \
    -o $output_path \
    -ldflags "-s -w -X cliaas.Version=${version}" \
    github.com/c0-ops/cliaas/cmd/cliaas
cd -
