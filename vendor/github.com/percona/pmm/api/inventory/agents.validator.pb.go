// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: inventory/agents.proto

package inventory

import fmt "fmt"
import github_com_mwitkow_go_proto_validators "github.com/mwitkow/go-proto-validators"
import proto "github.com/golang/protobuf/proto"
import math "math"
import _ "github.com/mwitkow/go-proto-validators"
import _ "google.golang.org/genproto/googleapis/api/annotations"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *PMMAgent) Validate() error {
	return nil
}
func (this *NodeExporter) Validate() error {
	return nil
}
func (this *MySQLdExporter) Validate() error {
	return nil
}
func (this *RDSExporter) Validate() error {
	return nil
}
func (this *ExternalAgent) Validate() error {
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
	for _, item := range this.RdsExporter {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("RdsExporter", err)
			}
		}
	}
	for _, item := range this.ExternalAgent {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("ExternalAgent", err)
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
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_RdsExporter); ok {
		if oneOfNester.RdsExporter != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.RdsExporter); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("RdsExporter", err)
			}
		}
	}
	if oneOfNester, ok := this.GetAgent().(*GetAgentResponse_ExternalAgent); ok {
		if oneOfNester.ExternalAgent != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.ExternalAgent); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("ExternalAgent", err)
			}
		}
	}
	return nil
}
func (this *AddPMMAgentRequest) Validate() error {
	if this.NodeId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("NodeId", fmt.Errorf(`value '%v' must not be an empty string`, this.NodeId))
	}
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
	if this.NodeId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("NodeId", fmt.Errorf(`value '%v' must not be an empty string`, this.NodeId))
	}
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
func (this *AddMySQLdExporterRequest) Validate() error {
	if this.RunsOnNodeId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("RunsOnNodeId", fmt.Errorf(`value '%v' must not be an empty string`, this.RunsOnNodeId))
	}
	if this.ServiceId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("ServiceId", fmt.Errorf(`value '%v' must not be an empty string`, this.ServiceId))
	}
	if this.Username == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Username", fmt.Errorf(`value '%v' must not be an empty string`, this.Username))
	}
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
func (this *AddRDSExporterRequest) Validate() error {
	if this.RunsOnNodeId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("RunsOnNodeId", fmt.Errorf(`value '%v' must not be an empty string`, this.RunsOnNodeId))
	}
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
func (this *AddExternalAgentRequest) Validate() error {
	return nil
}
func (this *AddExternalAgentResponse) Validate() error {
	if this.ExternalAgent != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.ExternalAgent); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("ExternalAgent", err)
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
