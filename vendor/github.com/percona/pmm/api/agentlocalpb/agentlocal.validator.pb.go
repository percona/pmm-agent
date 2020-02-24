// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: agentlocalpb/agentlocal.proto

package agentlocalpb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/golang/protobuf/ptypes/duration"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"
	github_com_mwitkow_go_proto_validators "github.com/mwitkow/go-proto-validators"
	_ "github.com/percona/pmm/api/inventorypb"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *ServerInfo) Validate() error {
	if this.Latency != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Latency); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Latency", err)
		}
	}
	if this.ClockDrift != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.ClockDrift); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("ClockDrift", err)
		}
	}
	return nil
}
func (this *AgentInfo) Validate() error {
	return nil
}
func (this *StatusRequest) Validate() error {
	return nil
}
func (this *StatusResponse) Validate() error {
	if this.ServerInfo != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.ServerInfo); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("ServerInfo", err)
		}
	}
	for _, item := range this.AgentsInfo {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("AgentsInfo", err)
			}
		}
	}
	return nil
}
func (this *ReloadRequest) Validate() error {
	return nil
}
func (this *ReloadResponse) Validate() error {
	return nil
}
