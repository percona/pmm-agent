// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: agentpb/agent.proto

package agentpb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/golang/protobuf/ptypes/duration"
	_ "github.com/golang/protobuf/ptypes/timestamp"
	github_com_mwitkow_go_proto_validators "github.com/mwitkow/go-proto-validators"
	_ "github.com/percona/pmm/api/inventorypb"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *Ping) Validate() error {
	return nil
}
func (this *Pong) Validate() error {
	if this.CurrentTime != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.CurrentTime); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("CurrentTime", err)
		}
	}
	return nil
}
func (this *QANCollectRequest) Validate() error {
	for _, item := range this.MetricsBucket {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("MetricsBucket", err)
			}
		}
	}
	return nil
}
func (this *QANCollectResponse) Validate() error {
	return nil
}
func (this *StateChangedRequest) Validate() error {
	return nil
}
func (this *StateChangedResponse) Validate() error {
	return nil
}
func (this *SetStateRequest) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *SetStateRequest_AgentProcess) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *SetStateRequest_BuiltinAgent) Validate() error {
	return nil
}
func (this *SetStateResponse) Validate() error {
	return nil
}
func (this *QueryActionValue) Validate() error {
	if oneOfNester, ok := this.GetKind().(*QueryActionValue_Timestamp); ok {
		if oneOfNester.Timestamp != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.Timestamp); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Timestamp", err)
			}
		}
	}
	if oneOfNester, ok := this.GetKind().(*QueryActionValue_Slice); ok {
		if oneOfNester.Slice != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.Slice); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Slice", err)
			}
		}
	}
	if oneOfNester, ok := this.GetKind().(*QueryActionValue_Map); ok {
		if oneOfNester.Map != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.Map); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Map", err)
			}
		}
	}
	return nil
}
func (this *QueryActionSlice) Validate() error {
	for _, item := range this.Slice {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Slice", err)
			}
		}
	}
	return nil
}
func (this *QueryActionMap) Validate() error {
	// Validation of proto3 map<> fields is unsupported.
	return nil
}
func (this *QueryActionResult) Validate() error {
	for _, item := range this.Res {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Res", err)
			}
		}
	}
	return nil
}
func (this *StartActionRequest) Validate() error {
	if oneOfNester, ok := this.GetParams().(*StartActionRequest_MysqlExplainParams); ok {
		if oneOfNester.MysqlExplainParams != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.MysqlExplainParams); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("MysqlExplainParams", err)
			}
		}
	}
	if oneOfNester, ok := this.GetParams().(*StartActionRequest_MysqlShowCreateTableParams); ok {
		if oneOfNester.MysqlShowCreateTableParams != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.MysqlShowCreateTableParams); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("MysqlShowCreateTableParams", err)
			}
		}
	}
	if oneOfNester, ok := this.GetParams().(*StartActionRequest_MysqlShowTableStatusParams); ok {
		if oneOfNester.MysqlShowTableStatusParams != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.MysqlShowTableStatusParams); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("MysqlShowTableStatusParams", err)
			}
		}
	}
	if oneOfNester, ok := this.GetParams().(*StartActionRequest_MysqlShowIndexParams); ok {
		if oneOfNester.MysqlShowIndexParams != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.MysqlShowIndexParams); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("MysqlShowIndexParams", err)
			}
		}
	}
	if oneOfNester, ok := this.GetParams().(*StartActionRequest_PostgresqlShowCreateTableParams); ok {
		if oneOfNester.PostgresqlShowCreateTableParams != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.PostgresqlShowCreateTableParams); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("PostgresqlShowCreateTableParams", err)
			}
		}
	}
	if oneOfNester, ok := this.GetParams().(*StartActionRequest_PostgresqlShowIndexParams); ok {
		if oneOfNester.PostgresqlShowIndexParams != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.PostgresqlShowIndexParams); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("PostgresqlShowIndexParams", err)
			}
		}
	}
	if oneOfNester, ok := this.GetParams().(*StartActionRequest_MongodbExplainParams); ok {
		if oneOfNester.MongodbExplainParams != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.MongodbExplainParams); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("MongodbExplainParams", err)
			}
		}
	}
	if oneOfNester, ok := this.GetParams().(*StartActionRequest_MysqlQueryShowParams); ok {
		if oneOfNester.MysqlQueryShowParams != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.MysqlQueryShowParams); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("MysqlQueryShowParams", err)
			}
		}
	}
	if oneOfNester, ok := this.GetParams().(*StartActionRequest_MysqlQuerySelectParams); ok {
		if oneOfNester.MysqlQuerySelectParams != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.MysqlQuerySelectParams); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("MysqlQuerySelectParams", err)
			}
		}
	}
	if oneOfNester, ok := this.GetParams().(*StartActionRequest_PostgresqlQueryShowParams); ok {
		if oneOfNester.PostgresqlQueryShowParams != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.PostgresqlQueryShowParams); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("PostgresqlQueryShowParams", err)
			}
		}
	}
	if oneOfNester, ok := this.GetParams().(*StartActionRequest_PostgresqlQuerySelectParams); ok {
		if oneOfNester.PostgresqlQuerySelectParams != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.PostgresqlQuerySelectParams); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("PostgresqlQuerySelectParams", err)
			}
		}
	}
	if oneOfNester, ok := this.GetParams().(*StartActionRequest_MongodbQueryGetparameterParams); ok {
		if oneOfNester.MongodbQueryGetparameterParams != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.MongodbQueryGetparameterParams); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("MongodbQueryGetparameterParams", err)
			}
		}
	}
	return nil
}
func (this *StartActionRequest_MySQLExplainParams) Validate() error {
	return nil
}
func (this *StartActionRequest_MySQLShowCreateTableParams) Validate() error {
	return nil
}
func (this *StartActionRequest_MySQLShowTableStatusParams) Validate() error {
	return nil
}
func (this *StartActionRequest_MySQLShowIndexParams) Validate() error {
	return nil
}
func (this *StartActionRequest_PostgreSQLShowCreateTableParams) Validate() error {
	return nil
}
func (this *StartActionRequest_PostgreSQLShowIndexParams) Validate() error {
	return nil
}
func (this *StartActionRequest_MongoDBExplainParams) Validate() error {
	return nil
}
func (this *StartActionRequest_MySQLQueryShowParams) Validate() error {
	return nil
}
func (this *StartActionRequest_MySQLQuerySelectParams) Validate() error {
	return nil
}
func (this *StartActionRequest_PostgreSQLQueryShowParams) Validate() error {
	return nil
}
func (this *StartActionRequest_PostgreSQLQuerySelectParams) Validate() error {
	return nil
}
func (this *StartActionRequest_MongoDBQueryGetParameterParams) Validate() error {
	return nil
}
func (this *StartActionResponse) Validate() error {
	return nil
}
func (this *StopActionRequest) Validate() error {
	return nil
}
func (this *StopActionResponse) Validate() error {
	return nil
}
func (this *ActionResultRequest) Validate() error {
	return nil
}
func (this *ActionResultResponse) Validate() error {
	return nil
}
func (this *CheckConnectionRequest) Validate() error {
	if this.Timeout != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Timeout); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Timeout", err)
		}
	}
	return nil
}
func (this *CheckConnectionResponse) Validate() error {
	if this.Stats != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Stats); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Stats", err)
		}
	}
	return nil
}
func (this *CheckConnectionResponse_Stats) Validate() error {
	return nil
}
func (this *AgentMessage) Validate() error {
	if oneOfNester, ok := this.GetPayload().(*AgentMessage_Ping); ok {
		if oneOfNester.Ping != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.Ping); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Ping", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*AgentMessage_StateChanged); ok {
		if oneOfNester.StateChanged != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.StateChanged); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("StateChanged", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*AgentMessage_QanCollect); ok {
		if oneOfNester.QanCollect != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.QanCollect); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("QanCollect", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*AgentMessage_ActionResult); ok {
		if oneOfNester.ActionResult != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.ActionResult); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("ActionResult", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*AgentMessage_Pong); ok {
		if oneOfNester.Pong != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.Pong); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Pong", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*AgentMessage_SetState); ok {
		if oneOfNester.SetState != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.SetState); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("SetState", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*AgentMessage_StartAction); ok {
		if oneOfNester.StartAction != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.StartAction); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("StartAction", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*AgentMessage_StopAction); ok {
		if oneOfNester.StopAction != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.StopAction); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("StopAction", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*AgentMessage_CheckConnection); ok {
		if oneOfNester.CheckConnection != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.CheckConnection); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("CheckConnection", err)
			}
		}
	}
	return nil
}
func (this *ServerMessage) Validate() error {
	if oneOfNester, ok := this.GetPayload().(*ServerMessage_Pong); ok {
		if oneOfNester.Pong != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.Pong); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Pong", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*ServerMessage_StateChanged); ok {
		if oneOfNester.StateChanged != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.StateChanged); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("StateChanged", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*ServerMessage_QanCollect); ok {
		if oneOfNester.QanCollect != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.QanCollect); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("QanCollect", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*ServerMessage_ActionResult); ok {
		if oneOfNester.ActionResult != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.ActionResult); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("ActionResult", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*ServerMessage_Ping); ok {
		if oneOfNester.Ping != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.Ping); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Ping", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*ServerMessage_SetState); ok {
		if oneOfNester.SetState != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.SetState); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("SetState", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*ServerMessage_StartAction); ok {
		if oneOfNester.StartAction != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.StartAction); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("StartAction", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*ServerMessage_StopAction); ok {
		if oneOfNester.StopAction != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.StopAction); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("StopAction", err)
			}
		}
	}
	if oneOfNester, ok := this.GetPayload().(*ServerMessage_CheckConnection); ok {
		if oneOfNester.CheckConnection != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.CheckConnection); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("CheckConnection", err)
			}
		}
	}
	return nil
}
