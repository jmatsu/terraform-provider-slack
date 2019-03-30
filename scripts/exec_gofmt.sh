#!/usr/bin/env bash

gofmt_files() {
    find . -name '*.go' | grep -v vendor
}

gofmt -w $(gofmt_files)