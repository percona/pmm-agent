help:                           ## Display this help message.
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
		awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

# `cut` is used to remove first `v` from `git describe` output
PMM_RELEASE_PATH ?= bin
PMM_RELEASE_VERSION ?= $(shell git describe --always --dirty | cut -b2-)
PMM_RELEASE_TIMESTAMP ?= $(shell date '+%s')
PMM_RELEASE_FULLCOMMIT ?= $(shell git rev-parse HEAD)
PMM_RELEASE_BRANCH ?= $(shell git describe --always --contains --all)
PMM_DEV_SERVER_PORT ?= 443

VERSION_FLAGS = -X 'github.com/percona/pmm-agent/vendor/github.com/percona/pmm/version.ProjectName=pmm-agent' \
				-X 'github.com/percona/pmm-agent/vendor/github.com/percona/pmm/version.Version=$(PMM_RELEASE_VERSION)' \
				-X 'github.com/percona/pmm-agent/vendor/github.com/percona/pmm/version.PMMVersion=$(PMM_RELEASE_VERSION)' \
				-X 'github.com/percona/pmm-agent/vendor/github.com/percona/pmm/version.Timestamp=$(PMM_RELEASE_TIMESTAMP)' \
				-X 'github.com/percona/pmm-agent/vendor/github.com/percona/pmm/version.FullCommit=$(PMM_RELEASE_FULLCOMMIT)' \
				-X 'github.com/percona/pmm-agent/vendor/github.com/percona/pmm/version.Branch=$(PMM_RELEASE_BRANCH)'

release:                        ## Build static pmm-agent release binary (Linux only).
	env CGO_ENABLED=1 go build -v -ldflags "-extldflags '-static' $(VERSION_FLAGS)" -tags 'osusergo netgo static_build' -o $(PMM_RELEASE_PATH)/pmm-agent
	$(PMM_RELEASE_PATH)/pmm-agent --version
	-ldd $(PMM_RELEASE_PATH)/pmm-agent

init:                           ## Installs tools to $GOPATH/bin (which is expected to be in $PATH).
	curl https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin

	go install ./vendor/github.com/BurntSushi/go-sumtype \
				./vendor/github.com/vektra/mockery/cmd/mockery \
				./vendor/golang.org/x/perf/cmd/benchstat \
				./vendor/golang.org/x/tools/cmd/goimports \
				./vendor/gopkg.in/reform.v1/reform

	go test -i ./...
	go test -race -i ./...

gen:                            ## Generate files.
	go generate ./...
	make format

# generate stub models for perfschema QAN agent; hidden from help as it is not generally useful
gen-init:
	go install ./vendor/gopkg.in/reform.v1/reform-db
	mkdir tmp-mysql
	reform-db -db-driver=mysql -db-source='root:root-password@tcp(127.0.0.1:3306)/performance_schema' init tmp-mysql

install:                        ## Install pmm-agent binary.
	go install -ldflags "$(VERSION_FLAGS)" ./...

install-race:                   ## Install pmm-agent binary with race detector.
	go install -ldflags "$(VERSION_FLAGS)" -race ./...

# TODO https://jira.percona.com/browse/PMM-4681
# TEST_PARALLEL_PACKAGES ?= foo bar
# go test $(TEST_FLAGS) $(TEST_PARALLEL_PACKAGES) - without `-p 1`

TEST_PACKAGES ?= ./...
TEST_FLAGS ?= -timeout=60s

test:                           ## Run tests.
	go test $(TEST_FLAGS) -p 1 $(TEST_PACKAGES)

test-race:                      ## Run tests with race detector.
	go test $(TEST_FLAGS) -p 1 -race $(TEST_PACKAGES)

test-cover:                     ## Run tests and collect per-package coverage information.
	go test $(TEST_FLAGS) -p 1 -coverprofile=cover.out -covermode=count $(TEST_PACKAGES)

test-crosscover:                ## Run tests and collect cross-package coverage information.
	go test $(TEST_FLAGS) -p 1 -coverprofile=crosscover.out -covermode=count -coverpkg=./... $(TEST_PACKAGES)

fuzz-slowlog-parser:            ## Run fuzzer for agents/mysql/slowlog/parser package.
	# go get -u github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-build
	mkdir -p agents/mysql/slowlog/parser/corpus
	cp agents/mysql/slowlog/parser/testdata/*.log agents/mysql/slowlog/parser/corpus/
	cd agents/mysql/slowlog/parser && go-fuzz-build
	cd agents/mysql/slowlog/parser && go-fuzz

fuzz-postgres-parser:   ## Run fuzzer for agents/postgres/parser package.
	# FIXME see https://github.com/AlekSi/go-fuzz/pull/1

	# go get -u github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-build
	mkdir -p agents/postgres/parser/fuzzing/corpus
	cp agents/postgres/parser/testdata/*.sql agents/postgres/parser/fuzzing/corpus/
	cd agents/postgres/parser && go-fuzz-build
	cd agents/postgres/parser && go-fuzz -workdir fuzzing

bench:                          ## Run benchmarks.
	go test -bench=. -benchtime=3s -count=5 -cpu=1 -timeout=20m -failfast github.com/percona/pmm-agent/agents/mysql/slowlog/parser | tee slowlog_parser_new.bench
	benchstat slowlog_parser_old.bench slowlog_parser_new.bench

	go test -bench=. -benchtime=3s -count=5 -cpu=1 -timeout=20m -failfast github.com/percona/pmm-agent/agents/postgres/parser | tee postgres_parser_new.bench
	benchstat postgres_parser_old.bench postgres_parser_new.bench

check:                          ## Run required checkers and linters.
	go run .github/check-license.go
	go-sumtype ./vendor/... ./...

check-all: check                ## Run golang ci linter to check new changes from master.
	golangci-lint run -c=.golangci.yml --new-from-rev=master

FILES = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

format:                         ## Format source code.
	gofmt -w -s $(FILES)
	goimports -local github.com/percona/pmm-agent -l -w $(FILES)

RUN_FLAGS = --config-file=pmm-agent-dev.yaml

run: install _run               ## Run pmm-agent.

run-race: install-race _run     ## Run pmm-agent with race detector.

run-race-cover: install-race    ## Run pmm-agent with race detector and collect coverage information.
	go test -coverpkg="github.com/percona/pmm-agent/..." \
			-tags maincover \
			-ldflags "$(VERSION_FLAGS)" \
			-race -c -o bin/pmm-agent.test
	bin/pmm-agent.test -test.coverprofile=cover.out -test.run=TestMainCover -- $(RUN_FLAGS)

_run:
	pmm-agent $(RUN_FLAGS)

ENV_UP_FLAGS ?= --force-recreate --renew-anon-volumes --remove-orphans

env-up:                         ## Start development environment.
	# to make slowlog rotation tests work
	rm -fr testdata
	mkdir -p testdata/mysql/slowlogs
	chmod -R 0777 testdata

	docker-compose up $(ENV_UP_FLAGS)

env-down:                       ## Stop development environment.
	docker-compose down --volumes --remove-orphans

setup-dev: install              ## Run pmm-agent setup in development environment.
	pmm-agent setup $(RUN_FLAGS) --server-insecure-tls --server-address=127.0.0.1:${PMM_DEV_SERVER_PORT} --server-username=admin --server-password=admin --paths-exporters_base=$(GOPATH)/bin --force

env-mysql:                      ## Run mysql client.
	docker exec -ti pmm-agent_mysql mysql --host=127.0.0.1 --user=root --password=root-password

env-mongo:                      ## Run mongo client.
	docker exec -ti pmm-agent_mongo mongo --username=root --password=root-password

env-psql:                       ## Run psql client.
	docker exec -ti pmm-agent_postgres env PGPASSWORD=pmm-agent-password psql --username=pmm-agent

env-sysbench-prepare:
	docker-compose exec --workdir=/sysbench/sysbench-tpcc sysbench ./tpcc.lua \
		--db-driver=pgsql --pgsql-host=postgres --pgsql-user=pmm-agent --pgsql-password=pmm-agent-password --pgsql-db=pmm-agent \
		--threads=1 --time=0 --report-interval=10 \
		--tables=1 --scale=10  --use_fk=0 --enable_purge=yes prepare

env-sysbench-run:
	docker-compose exec --workdir=/sysbench/sysbench-tpcc sysbench ./tpcc.lua \
		--db-driver=pgsql --pgsql-host=postgres --pgsql-user=pmm-agent --pgsql-password=pmm-agent-password --pgsql-db=pmm-agent \
		--threads=4 --time=0 --rate=10 --report-interval=10 --percentile=99 \
		--tables=1 --scale=10  --use_fk=0 --enable_purge=yes run

ci-reviewdog:                   ## Runs reviewdog checks.
	golangci-lint run -c=.golangci-required.yml --out-format=line-number | bin/reviewdog -f=golangci-lint -level=error -reporter=github-pr-check
	golangci-lint run -c=.golangci.yml --out-format=line-number | bin/reviewdog -f=golangci-lint -level=error -reporter=github-pr-review
