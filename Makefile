install:                  ## Install this program.
	go install -v ./...

exporters:                ## Install exporters.
	go install -v ./vendor/github.com/percona/mysqld_exporter

serve: install exporters  ## Start program as server and listen for incoming http requests.
	pmm-agent serve

test: exporters           ## Run tests.
	go test -mod=vendor -v -race ./...

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

verify:                   ## ensure that vendor/ is in sync with go.*
	go mod vendor
	git diff --exit-code

help: Makefile            ## Display this help message.
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
	    sort | \
	    awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

.DEFAULT_GOAL := help
.PHONY: install exporters serve test gen api lint format help
