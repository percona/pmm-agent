// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: inventorypb/agents.proto

package inventorypb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/mwitkow/go-proto-validators"
	github_com_mwitkow_go_proto_validators "github.com/mwitkow/go-proto-validators"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *PMMAgent) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *VMAgent) Validate() error {
	return nil
}
func (this *NodeExporter) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *MySQLdExporter) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *MongoDBExporter) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *PostgresExporter) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *ProxySQLExporter) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *QANMySQLPerfSchemaAgent) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *QANMySQLSlowlogAgent) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *QANMongoDBProfilerAgent) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *QANPostgreSQLPgStatementsAgent) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *QANPostgreSQLPgStatMonitorAgent) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *RDSExporter) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *ExternalExporter) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AzureExporter) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *ChangeCommonAgentParams) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *ListAgentsRequest) Validate() error {
	return nil
}
func (this *ListAgentsResponse) Validate() error {
	for _, item := range this.PmmAgent {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("PmmAgent", err)
			}
		}
	}
	for _, item := range this.VmAgent {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("VmAgent", err)
			}
		}
	}
	for _, item := range this.NodeExporter {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("NodeExporter", err)
			}
		}
	}
	for _, item := range this.MysqldExporter {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("MysqldExporter", err)
			}
		}
	}
	for _, item := range this.MongodbExporter {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("MongodbExporter", err)
			}
		}
	}
	for _, item := range this.PostgresExporter {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("PostgresExporter", err)
			}
		}
	}
	for _, item := range this.ProxysqlExporter {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("ProxysqlExporter", err)
			}
		}
	}
	for _, item := range this.QanMysqlPerfschemaAgent {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("QanMysqlPerfschemaAgent", err)
			}
		}
	}
	for _, item := range this.QanMysqlSlowlogAgent {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("QanMysqlSlowlogAgent", err)
			}
		}
	}
	for _, item := range this.QanMongodbProfilerAgent {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("QanMongodbProfilerAgent", err)
			}
		}
	}
	for _, item := range this.QanPostgresqlPgstatementsAgent {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("QanPostgresqlPgstatementsAgent", err)
			}
		}
	}
	for _, item := range this.QanPostgresqlPgstatmonitorAgent {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("QanPostgresqlPgstatmonitorAgent", err)
			}
		}
	}
	for _, item := range this.RdsExporter {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("RdsExporter", err)
			}
		}
	}
	for _, item := range this.ExternalExporter {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("ExternalExporter", err)
			}
		}
	}
	for _, item := range this.AzureExporter {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("AzureExporter", err)
			}
		}
	}
	return nil
}
func (this *GetAgentRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	return nil
}
func (this *GetAgentResponse) Validate() error {
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_PmmAgent); ok {
		if oneOfNester.PmmAgent != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.PmmAgent); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("PmmAgent", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_Vmagent); ok {
		if oneOfNester.Vmagent != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.Vmagent); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Vmagent", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_NodeExporter); ok {
		if oneOfNester.NodeExporter != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.NodeExporter); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("NodeExporter", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_MysqldExporter); ok {
		if oneOfNester.MysqldExporter != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.MysqldExporter); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("MysqldExporter", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_MongodbExporter); ok {
		if oneOfNester.MongodbExporter != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.MongodbExporter); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("MongodbExporter", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_PostgresExporter); ok {
		if oneOfNester.PostgresExporter != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.PostgresExporter); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("PostgresExporter", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_ProxysqlExporter); ok {
		if oneOfNester.ProxysqlExporter != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.ProxysqlExporter); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("ProxysqlExporter", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_QanMysqlPerfschemaAgent); ok {
		if oneOfNester.QanMysqlPerfschemaAgent != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.QanMysqlPerfschemaAgent); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("QanMysqlPerfschemaAgent", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_QanMysqlSlowlogAgent); ok {
		if oneOfNester.QanMysqlSlowlogAgent != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.QanMysqlSlowlogAgent); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("QanMysqlSlowlogAgent", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_QanMongodbProfilerAgent); ok {
		if oneOfNester.QanMongodbProfilerAgent != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.QanMongodbProfilerAgent); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("QanMongodbProfilerAgent", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_QanPostgresqlPgstatementsAgent); ok {
		if oneOfNester.QanPostgresqlPgstatementsAgent != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.QanPostgresqlPgstatementsAgent); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("QanPostgresqlPgstatementsAgent", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_QanPostgresqlPgstatmonitorAgent); ok {
		if oneOfNester.QanPostgresqlPgstatmonitorAgent != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.QanPostgresqlPgstatmonitorAgent); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("QanPostgresqlPgstatmonitorAgent", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_RdsExporter); ok {
		if oneOfNester.RdsExporter != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.RdsExporter); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("RdsExporter", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_ExternalExporter); ok {
		if oneOfNester.ExternalExporter != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.ExternalExporter); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("ExternalExporter", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_AzureExporter); ok {
		if oneOfNester.AzureExporter != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.AzureExporter); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("AzureExporter", err)
			}
		}
	}
	return nil
}
func (this *AddPMMAgentRequest) Validate() error {
	if this.RunsOnNodeId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("RunsOnNodeId", fmt.Errorf(`value '%v' must not be an empty string`, this.RunsOnNodeId))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddPMMAgentResponse) Validate() error {
	if this.PmmAgent != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.PmmAgent); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("PmmAgent", err)
		}
	}
	return nil
}
func (this *AddNodeExporterRequest) Validate() error {
	if this.PmmAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("PmmAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.PmmAgentId))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddNodeExporterResponse) Validate() error {
	if this.NodeExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.NodeExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("NodeExporter", err)
		}
	}
	return nil
}
func (this *ChangeNodeExporterRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	if this.Common != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Common); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Common", err)
		}
	}
	return nil
}
func (this *ChangeNodeExporterResponse) Validate() error {
	if this.NodeExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.NodeExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("NodeExporter", err)
		}
	}
	return nil
}
func (this *AddMySQLdExporterRequest) Validate() error {
	if this.PmmAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("PmmAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.PmmAgentId))
	}
	if this.ServiceId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("ServiceId", fmt.Errorf(`value '%v' must not be an empty string`, this.ServiceId))
	}
	if this.Username == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Username", fmt.Errorf(`value '%v' must not be an empty string`, this.Username))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddMySQLdExporterResponse) Validate() error {
	if this.MysqldExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.MysqldExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("MysqldExporter", err)
		}
	}
	return nil
}
func (this *ChangeMySQLdExporterRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	if this.Common != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Common); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Common", err)
		}
	}
	return nil
}
func (this *ChangeMySQLdExporterResponse) Validate() error {
	if this.MysqldExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.MysqldExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("MysqldExporter", err)
		}
	}
	return nil
}
func (this *AddMongoDBExporterRequest) Validate() error {
	if this.PmmAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("PmmAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.PmmAgentId))
	}
	if this.ServiceId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("ServiceId", fmt.Errorf(`value '%v' must not be an empty string`, this.ServiceId))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddMongoDBExporterResponse) Validate() error {
	if this.MongodbExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.MongodbExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("MongodbExporter", err)
		}
	}
	return nil
}
func (this *ChangeMongoDBExporterRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	if this.Common != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Common); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Common", err)
		}
	}
	return nil
}
func (this *ChangeMongoDBExporterResponse) Validate() error {
	if this.MongodbExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.MongodbExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("MongodbExporter", err)
		}
	}
	return nil
}
func (this *AddPostgresExporterRequest) Validate() error {
	if this.PmmAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("PmmAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.PmmAgentId))
	}
	if this.ServiceId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("ServiceId", fmt.Errorf(`value '%v' must not be an empty string`, this.ServiceId))
	}
	if this.Username == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Username", fmt.Errorf(`value '%v' must not be an empty string`, this.Username))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddPostgresExporterResponse) Validate() error {
	if this.PostgresExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.PostgresExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("PostgresExporter", err)
		}
	}
	return nil
}
func (this *ChangePostgresExporterRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	if this.Common != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Common); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Common", err)
		}
	}
	return nil
}
func (this *ChangePostgresExporterResponse) Validate() error {
	if this.PostgresExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.PostgresExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("PostgresExporter", err)
		}
	}
	return nil
}
func (this *AddProxySQLExporterRequest) Validate() error {
	if this.PmmAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("PmmAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.PmmAgentId))
	}
	if this.ServiceId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("ServiceId", fmt.Errorf(`value '%v' must not be an empty string`, this.ServiceId))
	}
	if this.Username == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Username", fmt.Errorf(`value '%v' must not be an empty string`, this.Username))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddProxySQLExporterResponse) Validate() error {
	if this.ProxysqlExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.ProxysqlExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("ProxysqlExporter", err)
		}
	}
	return nil
}
func (this *ChangeProxySQLExporterRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	if this.Common != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Common); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Common", err)
		}
	}
	return nil
}
func (this *ChangeProxySQLExporterResponse) Validate() error {
	if this.ProxysqlExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.ProxysqlExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("ProxysqlExporter", err)
		}
	}
	return nil
}
func (this *AddQANMySQLPerfSchemaAgentRequest) Validate() error {
	if this.PmmAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("PmmAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.PmmAgentId))
	}
	if this.ServiceId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("ServiceId", fmt.Errorf(`value '%v' must not be an empty string`, this.ServiceId))
	}
	if this.Username == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Username", fmt.Errorf(`value '%v' must not be an empty string`, this.Username))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddQANMySQLPerfSchemaAgentResponse) Validate() error {
	if this.QanMysqlPerfschemaAgent != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.QanMysqlPerfschemaAgent); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("QanMysqlPerfschemaAgent", err)
		}
	}
	return nil
}
func (this *ChangeQANMySQLPerfSchemaAgentRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	if this.Common != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Common); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Common", err)
		}
	}
	return nil
}
func (this *ChangeQANMySQLPerfSchemaAgentResponse) Validate() error {
	if this.QanMysqlPerfschemaAgent != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.QanMysqlPerfschemaAgent); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("QanMysqlPerfschemaAgent", err)
		}
	}
	return nil
}
func (this *AddQANMySQLSlowlogAgentRequest) Validate() error {
	if this.PmmAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("PmmAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.PmmAgentId))
	}
	if this.ServiceId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("ServiceId", fmt.Errorf(`value '%v' must not be an empty string`, this.ServiceId))
	}
	if this.Username == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Username", fmt.Errorf(`value '%v' must not be an empty string`, this.Username))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddQANMySQLSlowlogAgentResponse) Validate() error {
	if this.QanMysqlSlowlogAgent != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.QanMysqlSlowlogAgent); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("QanMysqlSlowlogAgent", err)
		}
	}
	return nil
}
func (this *ChangeQANMySQLSlowlogAgentRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	if this.Common != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Common); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Common", err)
		}
	}
	return nil
}
func (this *ChangeQANMySQLSlowlogAgentResponse) Validate() error {
	if this.QanMysqlSlowlogAgent != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.QanMysqlSlowlogAgent); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("QanMysqlSlowlogAgent", err)
		}
	}
	return nil
}
func (this *AddQANMongoDBProfilerAgentRequest) Validate() error {
	if this.PmmAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("PmmAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.PmmAgentId))
	}
	if this.ServiceId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("ServiceId", fmt.Errorf(`value '%v' must not be an empty string`, this.ServiceId))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddQANMongoDBProfilerAgentResponse) Validate() error {
	if this.QanMongodbProfilerAgent != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.QanMongodbProfilerAgent); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("QanMongodbProfilerAgent", err)
		}
	}
	return nil
}
func (this *ChangeQANMongoDBProfilerAgentRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	if this.Common != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Common); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Common", err)
		}
	}
	return nil
}
func (this *ChangeQANMongoDBProfilerAgentResponse) Validate() error {
	if this.QanMongodbProfilerAgent != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.QanMongodbProfilerAgent); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("QanMongodbProfilerAgent", err)
		}
	}
	return nil
}
func (this *AddQANPostgreSQLPgStatementsAgentRequest) Validate() error {
	if this.PmmAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("PmmAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.PmmAgentId))
	}
	if this.ServiceId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("ServiceId", fmt.Errorf(`value '%v' must not be an empty string`, this.ServiceId))
	}
	if this.Username == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Username", fmt.Errorf(`value '%v' must not be an empty string`, this.Username))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddQANPostgreSQLPgStatementsAgentResponse) Validate() error {
	if this.QanPostgresqlPgstatementsAgent != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.QanPostgresqlPgstatementsAgent); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("QanPostgresqlPgstatementsAgent", err)
		}
	}
	return nil
}
func (this *ChangeQANPostgreSQLPgStatementsAgentRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	if this.Common != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Common); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Common", err)
		}
	}
	return nil
}
func (this *ChangeQANPostgreSQLPgStatementsAgentResponse) Validate() error {
	if this.QanPostgresqlPgstatementsAgent != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.QanPostgresqlPgstatementsAgent); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("QanPostgresqlPgstatementsAgent", err)
		}
	}
	return nil
}
func (this *AddQANPostgreSQLPgStatMonitorAgentRequest) Validate() error {
	if this.PmmAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("PmmAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.PmmAgentId))
	}
	if this.ServiceId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("ServiceId", fmt.Errorf(`value '%v' must not be an empty string`, this.ServiceId))
	}
	if this.Username == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Username", fmt.Errorf(`value '%v' must not be an empty string`, this.Username))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddQANPostgreSQLPgStatMonitorAgentResponse) Validate() error {
	if this.QanPostgresqlPgstatmonitorAgent != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.QanPostgresqlPgstatmonitorAgent); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("QanPostgresqlPgstatmonitorAgent", err)
		}
	}
	return nil
}
func (this *ChangeQANPostgreSQLPgStatMonitorAgentRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	if this.Common != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Common); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Common", err)
		}
	}
	return nil
}
func (this *ChangeQANPostgreSQLPgStatMonitorAgentResponse) Validate() error {
	if this.QanPostgresqlPgstatmonitorAgent != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.QanPostgresqlPgstatmonitorAgent); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("QanPostgresqlPgstatmonitorAgent", err)
		}
	}
	return nil
}
func (this *AddRDSExporterRequest) Validate() error {
	if this.PmmAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("PmmAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.PmmAgentId))
	}
	if this.NodeId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("NodeId", fmt.Errorf(`value '%v' must not be an empty string`, this.NodeId))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddRDSExporterResponse) Validate() error {
	if this.RdsExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.RdsExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("RdsExporter", err)
		}
	}
	return nil
}
func (this *ChangeRDSExporterRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	if this.Common != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Common); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Common", err)
		}
	}
	return nil
}
func (this *ChangeRDSExporterResponse) Validate() error {
	if this.RdsExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.RdsExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("RdsExporter", err)
		}
	}
	return nil
}
func (this *AddExternalExporterRequest) Validate() error {
	if this.RunsOnNodeId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("RunsOnNodeId", fmt.Errorf(`value '%v' must not be an empty string`, this.RunsOnNodeId))
	}
	if !(this.ListenPort > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("ListenPort", fmt.Errorf(`value '%v' must be greater than '0'`, this.ListenPort))
	}
	if !(this.ListenPort < 65536) {
		return github_com_mwitkow_go_proto_validators.FieldError("ListenPort", fmt.Errorf(`value '%v' must be less than '65536'`, this.ListenPort))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddExternalExporterResponse) Validate() error {
	if this.ExternalExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.ExternalExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("ExternalExporter", err)
		}
	}
	return nil
}
func (this *ChangeExternalExporterRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	if this.Common != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Common); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Common", err)
		}
	}
	return nil
}
func (this *ChangeExternalExporterResponse) Validate() error {
	if this.ExternalExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.ExternalExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("ExternalExporter", err)
		}
	}
	return nil
}
func (this *AddAzureExporterRequest) Validate() error {
	if this.PmmAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("PmmAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.PmmAgentId))
	}
	if this.NodeId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("NodeId", fmt.Errorf(`value '%v' must not be an empty string`, this.NodeId))
	}
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *AddAzureExporterResponse) Validate() error {
	if this.AzureExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.AzureExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("AzureExporter", err)
		}
	}
	return nil
}
func (this *ChangeAzureExporterRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	if this.Common != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Common); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Common", err)
		}
	}
	return nil
}
func (this *ChangeAzureExporterResponse) Validate() error {
	if this.AzureExporter != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.AzureExporter); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("AzureExporter", err)
		}
	}
	return nil
}
func (this *RemoveAgentRequest) Validate() error {
	if this.AgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("AgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.AgentId))
	}
	return nil
}
func (this *RemoveAgentResponse) Validate() error {
	return nil
}
