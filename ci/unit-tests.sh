#! /bin/bash -eu

export GOPATH=$(pwd)

mkdir -p src/github.com/pivotal-cf/
ln -s $(pwd)/on-demand-service-broker-sdk src/github.com/pivotal-cf/
cd src/github.com/pivotal-cf/on-demand-service-broker-sdk

scripts/run-tests.sh
