module github.com/percona/pmm-agent

go 1.16

replace gopkg.in/alecthomas/kingpin.v2 => github.com/Percona-Lab/kingpin v2.2.6-percona+incompatible

replace github.com/lfittl/pg_query_go v1.0.2 => github.com/Percona-Lab/pg_query_go v1.0.1-0.20190723081422-3fc3af54a6f7

require (
	github.com/AlekSi/pointer v1.1.0
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/aws/aws-sdk-go v1.38.52 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/go-openapi/analysis v0.20.1 // indirect
	github.com/go-openapi/errors v0.20.0 // indirect
	github.com/go-openapi/runtime v0.19.28
	github.com/go-openapi/strfmt v0.20.1 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-openapi/validate v0.20.2 // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/klauspost/compress v1.12.3 // indirect
	github.com/lfittl/pg_query_go v1.0.2
	github.com/lib/pq v1.10.2
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/montanaflynn/stats v0.6.6 // indirect
	github.com/percona/exporter_shared v0.7.2
	github.com/percona/go-mysql v0.0.0-20210427141028-73d29c6da78c
	github.com/percona/percona-toolkit v3.2.1+incompatible
	github.com/percona/pmm v0.0.0-20210601163123-a8b03f81a484
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/prometheus/common v0.25.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/objx v0.2.0
	github.com/stretchr/testify v1.7.0
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	go.mongodb.org/mongo-driver v1.5.3
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5 // indirect
	golang.org/x/sys v0.0.0-20210601080250-7ecdf8ef093b
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c
	google.golang.org/grpc v1.38.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/reform.v1 v1.5.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)
