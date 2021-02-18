// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: inventorypb/tunnels.proto

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

func (this *Tunnel) Validate() error {
	return nil
}
func (this *ListTunnelsRequest) Validate() error {
	return nil
}
func (this *ListTunnelsResponse) Validate() error {
	for _, item := range this.Tunnel {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Tunnel", err)
			}
		}
	}
	return nil
}
func (this *AddTunnelRequest) Validate() error {
	if this.ListenAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("ListenAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.ListenAgentId))
	}
	if !(this.ListenPort > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("ListenPort", fmt.Errorf(`value '%v' must be greater than '0'`, this.ListenPort))
	}
	if !(this.ListenPort < 65536) {
		return github_com_mwitkow_go_proto_validators.FieldError("ListenPort", fmt.Errorf(`value '%v' must be less than '65536'`, this.ListenPort))
	}
	if this.ConnectAgentId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("ConnectAgentId", fmt.Errorf(`value '%v' must not be an empty string`, this.ConnectAgentId))
	}
	if !(this.ConnectPort > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("ConnectPort", fmt.Errorf(`value '%v' must be greater than '0'`, this.ConnectPort))
	}
	if !(this.ConnectPort < 65536) {
		return github_com_mwitkow_go_proto_validators.FieldError("ConnectPort", fmt.Errorf(`value '%v' must be less than '65536'`, this.ConnectPort))
	}
	return nil
}
func (this *AddTunnelResponse) Validate() error {
	return nil
}
func (this *RemoveTunnelRequest) Validate() error {
	if this.TunnelId == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("TunnelId", fmt.Errorf(`value '%v' must not be an empty string`, this.TunnelId))
	}
	return nil
}
func (this *RemoveTunnelResponse) Validate() error {
	return nil
}
