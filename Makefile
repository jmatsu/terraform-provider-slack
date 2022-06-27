base_version := $(shell git describe --tags --dirty | sed -e 's/^v//g')

build:
	go build -ldflags="-X main.version=$(base_version) -X main.commit=n/a" .

install-%: build
	mkdir -p ~/.terraform.d/plugins/github.com/jmatsu/slack/$(base_version)/${@:install-%=%}
	mv terraform-provider-slack ~/.terraform.d/plugins/github.com/jmatsu/slack/$(base_version)/${@:install-%=%}/terraform-provider-slack_v$(base_version)

fmt:
	gofmt -w .

fmt-check:
	gofmt .

generate:
	go generate

test:
	TF_ACC=1 go test ./... -timeout 120m -v