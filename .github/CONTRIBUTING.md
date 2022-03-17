# Contributing notes

## Pre-requirements: 
git, make, curl, go, gcc, docker, docker-compose, pmm-server

## Local setup  
Install exporters: 
* [node_exporter](https://github.com/percona/node_exporter)
* [mysqld_exporter](https://github.com/percona/mysqld_exporter)
* [rds_exporter](https://github.com/percona/rds_exporter)
* [postgres_exporter](https://github.com/percona/postgres_exporter)
* [mongodb_exporter](https://github.com/percona/mongodb_exporter)
* [proxysql_exporter](https://github.com/percona/proxysql_exporter)

Run `make download-exporters` to download all exporters

Run `make init` to install dependencies.

#### To run pmm-agent
Run [PMM-server](https://github.com/percona/pmm) docker container or [pmm-managed](https://github.com/percona/pmm-managed).  
Run `make setup-dev` to configure pmm-agent
Run `make run` to run pmm-agent
 

## Testing
Run `make env-up` to set-up environment.    
Run `make test` to run tests. 

## Code style
Before making PR, please run `make check-all` locally to run all checkers and linters.
