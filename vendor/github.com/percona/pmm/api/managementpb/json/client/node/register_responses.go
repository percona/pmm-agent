// Code generated by go-swagger; DO NOT EDIT.

package node

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"
)

// RegisterReader is a Reader for the Register structure.
type RegisterReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *RegisterReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewRegisterOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewRegisterDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewRegisterOK creates a RegisterOK with default headers values
func NewRegisterOK() *RegisterOK {
	return &RegisterOK{}
}

/*RegisterOK handles this case with default header values.

A successful response.
*/
type RegisterOK struct {
	Payload *RegisterOKBody
}

func (o *RegisterOK) Error() string {
	return fmt.Sprintf("[POST /v0/management/Node/Register][%d] registerOk  %+v", 200, o.Payload)
}

func (o *RegisterOK) GetPayload() *RegisterOKBody {
	return o.Payload
}

func (o *RegisterOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(RegisterOKBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewRegisterDefault creates a RegisterDefault with default headers values
func NewRegisterDefault(code int) *RegisterDefault {
	return &RegisterDefault{
		_statusCode: code,
	}
}

/*RegisterDefault handles this case with default header values.

An error response.
*/
type RegisterDefault struct {
	_statusCode int

	Payload *RegisterDefaultBody
}

// Code gets the status code for the register default response
func (o *RegisterDefault) Code() int {
	return o._statusCode
}

func (o *RegisterDefault) Error() string {
	return fmt.Sprintf("[POST /v0/management/Node/Register][%d] Register default  %+v", o._statusCode, o.Payload)
}

func (o *RegisterDefault) GetPayload() *RegisterDefaultBody {
	return o.Payload
}

func (o *RegisterDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(RegisterDefaultBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

/*RegisterBody register body
swagger:model RegisterBody
*/
type RegisterBody struct {

	// Address FIXME https://jira.percona.com/browse/PMM-3786
	Address string `json:"address,omitempty"`

	// Node availability zone.
	Az string `json:"az,omitempty"`

	// Container identifier. If specified, must be a unique Docker container identifier.
	ContainerID string `json:"container_id,omitempty"`

	// Container name.
	ContainerName string `json:"container_name,omitempty"`

	// Custom user-assigned labels.
	CustomLabels map[string]string `json:"custom_labels,omitempty"`

	// Linux distribution name and version.
	Distro string `json:"distro,omitempty"`

	// Linux machine-id.
	// Must be unique across all Generic Nodes if specified.
	MachineID string `json:"machine_id,omitempty"`

	// Node model.
	NodeModel string `json:"node_model,omitempty"`

	// Unique across all Nodes user-defined name. Can't be changed.
	NodeName string `json:"node_name,omitempty"`

	// NodeType describes supported Node types.
	// Enum: [NODE_TYPE_INVALID GENERIC_NODE CONTAINER_NODE REMOTE_NODE REMOTE_AMAZON_RDS_NODE]
	NodeType *string `json:"node_type,omitempty"`

	// Node region.
	Region string `json:"region,omitempty"`

	// If true, and Node with that name already exist, it will be removed with all dependent Services and Agents.
	Reregister bool `json:"reregister,omitempty"`
}

// Validate validates this register body
func (o *RegisterBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.validateNodeType(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var registerBodyTypeNodeTypePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["NODE_TYPE_INVALID","GENERIC_NODE","CONTAINER_NODE","REMOTE_NODE","REMOTE_AMAZON_RDS_NODE"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		registerBodyTypeNodeTypePropEnum = append(registerBodyTypeNodeTypePropEnum, v)
	}
}

const (

	// RegisterBodyNodeTypeNODETYPEINVALID captures enum value "NODE_TYPE_INVALID"
	RegisterBodyNodeTypeNODETYPEINVALID string = "NODE_TYPE_INVALID"

	// RegisterBodyNodeTypeGENERICNODE captures enum value "GENERIC_NODE"
	RegisterBodyNodeTypeGENERICNODE string = "GENERIC_NODE"

	// RegisterBodyNodeTypeCONTAINERNODE captures enum value "CONTAINER_NODE"
	RegisterBodyNodeTypeCONTAINERNODE string = "CONTAINER_NODE"

	// RegisterBodyNodeTypeREMOTENODE captures enum value "REMOTE_NODE"
	RegisterBodyNodeTypeREMOTENODE string = "REMOTE_NODE"

	// RegisterBodyNodeTypeREMOTEAMAZONRDSNODE captures enum value "REMOTE_AMAZON_RDS_NODE"
	RegisterBodyNodeTypeREMOTEAMAZONRDSNODE string = "REMOTE_AMAZON_RDS_NODE"
)

// prop value enum
func (o *RegisterBody) validateNodeTypeEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, registerBodyTypeNodeTypePropEnum); err != nil {
		return err
	}
	return nil
}

func (o *RegisterBody) validateNodeType(formats strfmt.Registry) error {

	if swag.IsZero(o.NodeType) { // not required
		return nil
	}

	// value enum
	if err := o.validateNodeTypeEnum("body"+"."+"node_type", "body", *o.NodeType); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (o *RegisterBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RegisterBody) UnmarshalBinary(b []byte) error {
	var res RegisterBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*RegisterDefaultBody ErrorResponse is a message returned on HTTP error.
swagger:model RegisterDefaultBody
*/
type RegisterDefaultBody struct {

	// code
	Code int32 `json:"code,omitempty"`

	// error
	Error string `json:"error,omitempty"`

	// message
	Message string `json:"message,omitempty"`
}

// Validate validates this register default body
func (o *RegisterDefaultBody) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *RegisterDefaultBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RegisterDefaultBody) UnmarshalBinary(b []byte) error {
	var res RegisterDefaultBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*RegisterOKBody register OK body
swagger:model RegisterOKBody
*/
type RegisterOKBody struct {

	// container node
	ContainerNode *RegisterOKBodyContainerNode `json:"container_node,omitempty"`

	// generic node
	GenericNode *RegisterOKBodyGenericNode `json:"generic_node,omitempty"`

	// pmm agent
	PMMAgent *RegisterOKBodyPMMAgent `json:"pmm_agent,omitempty"`
}

// Validate validates this register OK body
func (o *RegisterOKBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.validateContainerNode(formats); err != nil {
		res = append(res, err)
	}

	if err := o.validateGenericNode(formats); err != nil {
		res = append(res, err)
	}

	if err := o.validatePMMAgent(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *RegisterOKBody) validateContainerNode(formats strfmt.Registry) error {

	if swag.IsZero(o.ContainerNode) { // not required
		return nil
	}

	if o.ContainerNode != nil {
		if err := o.ContainerNode.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("registerOk" + "." + "container_node")
			}
			return err
		}
	}

	return nil
}

func (o *RegisterOKBody) validateGenericNode(formats strfmt.Registry) error {

	if swag.IsZero(o.GenericNode) { // not required
		return nil
	}

	if o.GenericNode != nil {
		if err := o.GenericNode.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("registerOk" + "." + "generic_node")
			}
			return err
		}
	}

	return nil
}

func (o *RegisterOKBody) validatePMMAgent(formats strfmt.Registry) error {

	if swag.IsZero(o.PMMAgent) { // not required
		return nil
	}

	if o.PMMAgent != nil {
		if err := o.PMMAgent.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("registerOk" + "." + "pmm_agent")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (o *RegisterOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RegisterOKBody) UnmarshalBinary(b []byte) error {
	var res RegisterOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*RegisterOKBodyContainerNode ContainerNode represents a Docker container.
swagger:model RegisterOKBodyContainerNode
*/
type RegisterOKBodyContainerNode struct {

	// Address FIXME https://jira.percona.com/browse/PMM-3786
	Address string `json:"address,omitempty"`

	// Node availability zone. Auto-detected and auto-updated.
	Az string `json:"az,omitempty"`

	// Container identifier. If specified, must be a unique Docker container identifier.
	// Auto-detected and auto-updated.
	ContainerID string `json:"container_id,omitempty"`

	// Container name. Auto-detected and auto-updated.
	ContainerName string `json:"container_name,omitempty"`

	// Custom user-assigned labels.
	CustomLabels map[string]string `json:"custom_labels,omitempty"`

	// Linux machine-id of the Generic Node where this Container Node runs. Auto-detected and auto-updated.
	// If defined, Generic Node with that machine_id must exist.
	MachineID string `json:"machine_id,omitempty"`

	// Unique randomly generated instance identifier. Can't be changed.
	NodeID string `json:"node_id,omitempty"`

	// Node model. Auto-detected and auto-updated.
	NodeModel string `json:"node_model,omitempty"`

	// Unique across all Nodes user-defined name. Can't be changed.
	NodeName string `json:"node_name,omitempty"`

	// Node region. Auto-detected and auto-updated.
	Region string `json:"region,omitempty"`
}

// Validate validates this register OK body container node
func (o *RegisterOKBodyContainerNode) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *RegisterOKBodyContainerNode) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RegisterOKBodyContainerNode) UnmarshalBinary(b []byte) error {
	var res RegisterOKBodyContainerNode
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*RegisterOKBodyGenericNode GenericNode represents a bare metal server or virtual machine.
swagger:model RegisterOKBodyGenericNode
*/
type RegisterOKBodyGenericNode struct {

	// Address FIXME https://jira.percona.com/browse/PMM-3786
	Address string `json:"address,omitempty"`

	// Node availability zone. Auto-detected and auto-updated.
	Az string `json:"az,omitempty"`

	// Custom user-assigned labels. Can be changed.
	CustomLabels map[string]string `json:"custom_labels,omitempty"`

	// Linux distribution name and version. Auto-detected and auto-updated.
	Distro string `json:"distro,omitempty"`

	// Linux machine-id. Auto-detected and auto-updated.
	// Must be unique across all Generic Nodes if specified.
	MachineID string `json:"machine_id,omitempty"`

	// Unique randomly generated instance identifier. Can't be changed.
	NodeID string `json:"node_id,omitempty"`

	// Node model. Auto-detected and auto-updated.
	NodeModel string `json:"node_model,omitempty"`

	// Unique across all Nodes user-defined name. Can't be changed.
	NodeName string `json:"node_name,omitempty"`

	// Node region. Auto-detected and auto-updated.
	Region string `json:"region,omitempty"`
}

// Validate validates this register OK body generic node
func (o *RegisterOKBodyGenericNode) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *RegisterOKBodyGenericNode) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RegisterOKBodyGenericNode) UnmarshalBinary(b []byte) error {
	var res RegisterOKBodyGenericNode
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*RegisterOKBodyPMMAgent PMMAgent runs on Generic on Container Node.
swagger:model RegisterOKBodyPMMAgent
*/
type RegisterOKBodyPMMAgent struct {

	// Unique randomly generated instance identifier.
	AgentID string `json:"agent_id,omitempty"`

	// True if Agent is running and connected to pmm-managed.
	Connected bool `json:"connected,omitempty"`

	// Custom user-assigned labels.
	CustomLabels map[string]string `json:"custom_labels,omitempty"`

	// Node identifier where this instance runs.
	RunsOnNodeID string `json:"runs_on_node_id,omitempty"`
}

// Validate validates this register OK body PMM agent
func (o *RegisterOKBodyPMMAgent) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *RegisterOKBodyPMMAgent) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RegisterOKBodyPMMAgent) UnmarshalBinary(b []byte) error {
	var res RegisterOKBodyPMMAgent
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
