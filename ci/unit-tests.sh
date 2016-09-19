#! /bin/bash -eu

export GOPATH=$(pwd)
export PATH=$GOPATH/bin:$PATH

mkdir -p src/github.com/pivotal-cf/
ln -s $(pwd)/on-demand-services-sdk src/github.com/pivotal-cf/
cd src/github.com/pivotal-cf/on-demand-services-sdk

go get -v github.com/tools/godep
godep restore
scripts/run-tests.sh
