export GO111MODULE=on

install:                  ## Install this program.
	go install -v ./...

exporters:                ## Install exporters.
	go install -v ./vendor/github.com/percona/mysqld_exporter

serve: install exporters  ## Start program as server and listen for incoming http requests.
	pmm-agent serve

test: exporters           ## Run tests.
	go test -mod=vendor -v -race ./...

test-cover: exporters
	go install -v ./vendor/github.com/AlekSi/gocoverutil
	gocoverutil test -v -race ./...

send-cover: SHELL:=/bin/bash
send-cover:
	bash <(curl -s https://codecov.io/bash) -X fix

gen:                      ## Run `go generate`.
	go generate ./...

api:                      ## Generate api.
	go install -v ./vendor/github.com/golang/protobuf/protoc-gen-go
	protoc -Iapi api/*.proto --go_out=plugins=grpc:api

lint:	                  ## Run `golangci-lint`.
	golangci-lint run

format:	                  ## Run `goimports`.
	go install -v ./vendor/golang.org/x/tools/cmd/goimports
	goimports -local github.com/percona/pmm-agent -l -w $(shell find . -type f -name '*.go' -not -path "./vendor/*")

verify:                   ## Ensure that vendor/ is in sync with `go.*`.
	go mod verify
	go mod vendor
	git diff --exit-code

help: Makefile            ## Display this help message.
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
	    sort | \
	    awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

.DEFAULT_GOAL := help
.PHONY: install exporters serve test test-cover send-cover gen api lint format verify help
