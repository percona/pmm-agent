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

// RegisterNodeReader is a Reader for the RegisterNode structure.
type RegisterNodeReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *RegisterNodeReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewRegisterNodeOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewRegisterNodeDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewRegisterNodeOK creates a RegisterNodeOK with default headers values
func NewRegisterNodeOK() *RegisterNodeOK {
	return &RegisterNodeOK{}
}

/*RegisterNodeOK handles this case with default header values.

A successful response.
*/
type RegisterNodeOK struct {
	Payload *RegisterNodeOKBody
}

func (o *RegisterNodeOK) Error() string {
	return fmt.Sprintf("[POST /v1/management/Node/Register][%d] registerNodeOk  %+v", 200, o.Payload)
}

func (o *RegisterNodeOK) GetPayload() *RegisterNodeOKBody {
	return o.Payload
}

func (o *RegisterNodeOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(RegisterNodeOKBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewRegisterNodeDefault creates a RegisterNodeDefault with default headers values
func NewRegisterNodeDefault(code int) *RegisterNodeDefault {
	return &RegisterNodeDefault{
		_statusCode: code,
	}
}

/*RegisterNodeDefault handles this case with default header values.

An error response.
*/
type RegisterNodeDefault struct {
	_statusCode int

	Payload *RegisterNodeDefaultBody
}

// Code gets the status code for the register node default response
func (o *RegisterNodeDefault) Code() int {
	return o._statusCode
}

func (o *RegisterNodeDefault) Error() string {
	return fmt.Sprintf("[POST /v1/management/Node/Register][%d] RegisterNode default  %+v", o._statusCode, o.Payload)
}

func (o *RegisterNodeDefault) GetPayload() *RegisterNodeDefaultBody {
	return o.Payload
}

func (o *RegisterNodeDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(RegisterNodeDefaultBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

/*RegisterNodeBody register node body
swagger:model RegisterNodeBody
*/
type RegisterNodeBody struct {

	// NodeType describes supported Node types.
	// Enum: [NODE_TYPE_INVALID GENERIC_NODE CONTAINER_NODE REMOTE_NODE REMOTE_RDS_NODE]
	NodeType *string `json:"node_type,omitempty"`

	// Unique across all Nodes user-defined name.
	NodeName string `json:"node_name,omitempty"`

	// Node address (DNS name or IP).
	Address string `json:"address,omitempty"`

	// Linux machine-id.
	MachineID string `json:"machine_id,omitempty"`

	// Linux distribution name and version.
	Distro string `json:"distro,omitempty"`

	// Container identifier. If specified, must be a unique Docker container identifier.
	ContainerID string `json:"container_id,omitempty"`

	// Container name.
	ContainerName string `json:"container_name,omitempty"`

	// Node model.
	NodeModel string `json:"node_model,omitempty"`

	// Node region.
	Region string `json:"region,omitempty"`

	// Node availability zone.
	Az string `json:"az,omitempty"`

	// Custom user-assigned labels for Node.
	CustomLabels map[string]string `json:"custom_labels,omitempty"`

	// If true, and Node with that name already exist, it will be removed with all dependent Services and Agents.
	Reregister bool `json:"reregister,omitempty"`
}

// Validate validates this register node body
func (o *RegisterNodeBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.validateNodeType(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var registerNodeBodyTypeNodeTypePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["NODE_TYPE_INVALID","GENERIC_NODE","CONTAINER_NODE","REMOTE_NODE","REMOTE_RDS_NODE"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		registerNodeBodyTypeNodeTypePropEnum = append(registerNodeBodyTypeNodeTypePropEnum, v)
	}
}

const (

	// RegisterNodeBodyNodeTypeNODETYPEINVALID captures enum value "NODE_TYPE_INVALID"
	RegisterNodeBodyNodeTypeNODETYPEINVALID string = "NODE_TYPE_INVALID"

	// RegisterNodeBodyNodeTypeGENERICNODE captures enum value "GENERIC_NODE"
	RegisterNodeBodyNodeTypeGENERICNODE string = "GENERIC_NODE"

	// RegisterNodeBodyNodeTypeCONTAINERNODE captures enum value "CONTAINER_NODE"
	RegisterNodeBodyNodeTypeCONTAINERNODE string = "CONTAINER_NODE"

	// RegisterNodeBodyNodeTypeREMOTENODE captures enum value "REMOTE_NODE"
	RegisterNodeBodyNodeTypeREMOTENODE string = "REMOTE_NODE"

	// RegisterNodeBodyNodeTypeREMOTERDSNODE captures enum value "REMOTE_RDS_NODE"
	RegisterNodeBodyNodeTypeREMOTERDSNODE string = "REMOTE_RDS_NODE"
)

// prop value enum
func (o *RegisterNodeBody) validateNodeTypeEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, registerNodeBodyTypeNodeTypePropEnum); err != nil {
		return err
	}
	return nil
}

func (o *RegisterNodeBody) validateNodeType(formats strfmt.Registry) error {

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
func (o *RegisterNodeBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RegisterNodeBody) UnmarshalBinary(b []byte) error {
	var res RegisterNodeBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*RegisterNodeDefaultBody ErrorResponse is a message returned on HTTP error.
swagger:model RegisterNodeDefaultBody
*/
type RegisterNodeDefaultBody struct {

	// code
	Code int32 `json:"code,omitempty"`

	// error
	Error string `json:"error,omitempty"`

	// message
	Message string `json:"message,omitempty"`
}

// Validate validates this register node default body
func (o *RegisterNodeDefaultBody) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *RegisterNodeDefaultBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RegisterNodeDefaultBody) UnmarshalBinary(b []byte) error {
	var res RegisterNodeDefaultBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*RegisterNodeOKBody register node OK body
swagger:model RegisterNodeOKBody
*/
type RegisterNodeOKBody struct {

	// container node
	ContainerNode *RegisterNodeOKBodyContainerNode `json:"container_node,omitempty"`

	// generic node
	GenericNode *RegisterNodeOKBodyGenericNode `json:"generic_node,omitempty"`

	// pmm agent
	PMMAgent *RegisterNodeOKBodyPMMAgent `json:"pmm_agent,omitempty"`
}

// Validate validates this register node OK body
func (o *RegisterNodeOKBody) Validate(formats strfmt.Registry) error {
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

func (o *RegisterNodeOKBody) validateContainerNode(formats strfmt.Registry) error {

	if swag.IsZero(o.ContainerNode) { // not required
		return nil
	}

	if o.ContainerNode != nil {
		if err := o.ContainerNode.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("registerNodeOk" + "." + "container_node")
			}
			return err
		}
	}

	return nil
}

func (o *RegisterNodeOKBody) validateGenericNode(formats strfmt.Registry) error {

	if swag.IsZero(o.GenericNode) { // not required
		return nil
	}

	if o.GenericNode != nil {
		if err := o.GenericNode.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("registerNodeOk" + "." + "generic_node")
			}
			return err
		}
	}

	return nil
}

func (o *RegisterNodeOKBody) validatePMMAgent(formats strfmt.Registry) error {

	if swag.IsZero(o.PMMAgent) { // not required
		return nil
	}

	if o.PMMAgent != nil {
		if err := o.PMMAgent.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("registerNodeOk" + "." + "pmm_agent")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (o *RegisterNodeOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RegisterNodeOKBody) UnmarshalBinary(b []byte) error {
	var res RegisterNodeOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*RegisterNodeOKBodyContainerNode ContainerNode represents a Docker container.
swagger:model RegisterNodeOKBodyContainerNode
*/
type RegisterNodeOKBodyContainerNode struct {

	// Unique randomly generated instance identifier.
	NodeID string `json:"node_id,omitempty"`

	// Unique across all Nodes user-defined name.
	NodeName string `json:"node_name,omitempty"`

	// Node address (DNS name or IP).
	Address string `json:"address,omitempty"`

	// Linux machine-id of the Generic Node where this Container Node runs.
	MachineID string `json:"machine_id,omitempty"`

	// Container identifier. If specified, must be a unique Docker container identifier.
	ContainerID string `json:"container_id,omitempty"`

	// Container name.
	ContainerName string `json:"container_name,omitempty"`

	// Node model.
	NodeModel string `json:"node_model,omitempty"`

	// Node region.
	Region string `json:"region,omitempty"`

	// Node availability zone.
	Az string `json:"az,omitempty"`

	// Custom user-assigned labels.
	CustomLabels map[string]string `json:"custom_labels,omitempty"`
}

// Validate validates this register node OK body container node
func (o *RegisterNodeOKBodyContainerNode) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *RegisterNodeOKBodyContainerNode) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RegisterNodeOKBodyContainerNode) UnmarshalBinary(b []byte) error {
	var res RegisterNodeOKBodyContainerNode
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*RegisterNodeOKBodyGenericNode GenericNode represents a bare metal server or virtual machine.
swagger:model RegisterNodeOKBodyGenericNode
*/
type RegisterNodeOKBodyGenericNode struct {

	// Unique randomly generated instance identifier.
	NodeID string `json:"node_id,omitempty"`

	// Unique across all Nodes user-defined name.
	NodeName string `json:"node_name,omitempty"`

	// Node address (DNS name or IP).
	Address string `json:"address,omitempty"`

	// Linux machine-id.
	MachineID string `json:"machine_id,omitempty"`

	// Linux distribution name and version.
	Distro string `json:"distro,omitempty"`

	// Node model.
	NodeModel string `json:"node_model,omitempty"`

	// Node region.
	Region string `json:"region,omitempty"`

	// Node availability zone.
	Az string `json:"az,omitempty"`

	// Custom user-assigned labels.
	CustomLabels map[string]string `json:"custom_labels,omitempty"`
}

// Validate validates this register node OK body generic node
func (o *RegisterNodeOKBodyGenericNode) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *RegisterNodeOKBodyGenericNode) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RegisterNodeOKBodyGenericNode) UnmarshalBinary(b []byte) error {
	var res RegisterNodeOKBodyGenericNode
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*RegisterNodeOKBodyPMMAgent PMMAgent runs on Generic or Container Node.
swagger:model RegisterNodeOKBodyPMMAgent
*/
type RegisterNodeOKBodyPMMAgent struct {

	// Unique randomly generated instance identifier.
	AgentID string `json:"agent_id,omitempty"`

	// Node identifier where this instance runs.
	RunsOnNodeID string `json:"runs_on_node_id,omitempty"`

	// Custom user-assigned labels.
	CustomLabels map[string]string `json:"custom_labels,omitempty"`

	// True if Agent is running and connected to pmm-managed.
	Connected bool `json:"connected,omitempty"`
}

// Validate validates this register node OK body PMM agent
func (o *RegisterNodeOKBodyPMMAgent) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *RegisterNodeOKBodyPMMAgent) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *RegisterNodeOKBodyPMMAgent) UnmarshalBinary(b []byte) error {
	var res RegisterNodeOKBodyPMMAgent
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
