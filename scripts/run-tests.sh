#! /usr/bin/env bash

set -eu

find . -name \*_test.go | grep -v /vendor/ | xargs -n 1 dirname | sort -u | xargs ginkgo
