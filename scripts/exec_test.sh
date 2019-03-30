#!/usr/bin/env bash

test_files() {
    go list ./... | grep -v vendor
}

TF_ACC=1 go test $(test_files) -v "$@" -timeout 120m