// Code generated by go-swagger; DO NOT EDIT.

package postgre_sql

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

// AddPostgreSQLReader is a Reader for the AddPostgreSQL structure.
type AddPostgreSQLReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *AddPostgreSQLReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewAddPostgreSQLOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewAddPostgreSQLDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewAddPostgreSQLOK creates a AddPostgreSQLOK with default headers values
func NewAddPostgreSQLOK() *AddPostgreSQLOK {
	return &AddPostgreSQLOK{}
}

/*AddPostgreSQLOK handles this case with default header values.

A successful response.
*/
type AddPostgreSQLOK struct {
	Payload *AddPostgreSQLOKBody
}

func (o *AddPostgreSQLOK) Error() string {
	return fmt.Sprintf("[POST /v0/management/PostgreSQL/Add][%d] addPostgreSqlOk  %+v", 200, o.Payload)
}

func (o *AddPostgreSQLOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(AddPostgreSQLOKBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddPostgreSQLDefault creates a AddPostgreSQLDefault with default headers values
func NewAddPostgreSQLDefault(code int) *AddPostgreSQLDefault {
	return &AddPostgreSQLDefault{
		_statusCode: code,
	}
}

/*AddPostgreSQLDefault handles this case with default header values.

An error response.
*/
type AddPostgreSQLDefault struct {
	_statusCode int

	Payload *AddPostgreSQLDefaultBody
}

// Code gets the status code for the add postgre SQL default response
func (o *AddPostgreSQLDefault) Code() int {
	return o._statusCode
}

func (o *AddPostgreSQLDefault) Error() string {
	return fmt.Sprintf("[POST /v0/management/PostgreSQL/Add][%d] AddPostgreSQL default  %+v", o._statusCode, o.Payload)
}

func (o *AddPostgreSQLDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(AddPostgreSQLDefaultBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

/*AddPostgreSQLBody add postgre SQL body
swagger:model AddPostgreSQLBody
*/
type AddPostgreSQLBody struct {

	// Node and Service access address (DNS name or IP). Required.
	Address string `json:"address,omitempty"`

	// Custom user-assigned labels.
	CustomLabels map[string]string `json:"custom_labels,omitempty"`

	// Environment name.
	Environment string `json:"environment,omitempty"`

	// Node identifier on which a service is been running. Required.
	NodeID string `json:"node_id,omitempty"`

	// PostgreSQL password for scraping metrics.
	Password string `json:"password,omitempty"`

	// The "pmm-agent" identifier which should run agents. Required.
	PMMAgentID string `json:"pmm_agent_id,omitempty"`

	// Service Access port. Required.
	Port int64 `json:"port,omitempty"`

	// Unique across all Services user-defined name. Required.
	ServiceName string `json:"service_name,omitempty"`

	// Skip connection check.
	SkipConnectionCheck bool `json:"skip_connection_check,omitempty"`

	// PostgreSQL username for scraping metrics.
	Username string `json:"username,omitempty"`
}

// Validate validates this add postgre SQL body
func (o *AddPostgreSQLBody) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *AddPostgreSQLBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *AddPostgreSQLBody) UnmarshalBinary(b []byte) error {
	var res AddPostgreSQLBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*AddPostgreSQLDefaultBody ErrorResponse is a message returned on HTTP error.
swagger:model AddPostgreSQLDefaultBody
*/
type AddPostgreSQLDefaultBody struct {

	// code
	Code int32 `json:"code,omitempty"`

	// error
	Error string `json:"error,omitempty"`

	// message
	Message string `json:"message,omitempty"`
}

// Validate validates this add postgre SQL default body
func (o *AddPostgreSQLDefaultBody) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *AddPostgreSQLDefaultBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *AddPostgreSQLDefaultBody) UnmarshalBinary(b []byte) error {
	var res AddPostgreSQLDefaultBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*AddPostgreSQLOKBody add postgre SQL OK body
swagger:model AddPostgreSQLOKBody
*/
type AddPostgreSQLOKBody struct {

	// postgres exporter
	PostgresExporter *AddPostgreSQLOKBodyPostgresExporter `json:"postgres_exporter,omitempty"`

	// service
	Service *AddPostgreSQLOKBodyService `json:"service,omitempty"`
}

// Validate validates this add postgre SQL OK body
func (o *AddPostgreSQLOKBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.validatePostgresExporter(formats); err != nil {
		res = append(res, err)
	}

	if err := o.validateService(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *AddPostgreSQLOKBody) validatePostgresExporter(formats strfmt.Registry) error {

	if swag.IsZero(o.PostgresExporter) { // not required
		return nil
	}

	if o.PostgresExporter != nil {
		if err := o.PostgresExporter.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("addPostgreSqlOk" + "." + "postgres_exporter")
			}
			return err
		}
	}

	return nil
}

func (o *AddPostgreSQLOKBody) validateService(formats strfmt.Registry) error {

	if swag.IsZero(o.Service) { // not required
		return nil
	}

	if o.Service != nil {
		if err := o.Service.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("addPostgreSqlOk" + "." + "service")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (o *AddPostgreSQLOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *AddPostgreSQLOKBody) UnmarshalBinary(b []byte) error {
	var res AddPostgreSQLOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*AddPostgreSQLOKBodyPostgresExporter PostgresExporter runs on Generic or Container Node and exposes PostgreSQL Service metrics.
swagger:model AddPostgreSQLOKBodyPostgresExporter
*/
type AddPostgreSQLOKBodyPostgresExporter struct {

	// Unique randomly generated instance identifier.
	AgentID string `json:"agent_id,omitempty"`

	// Custom user-assigned labels.
	CustomLabels map[string]string `json:"custom_labels,omitempty"`

	// Desired Agent status: enabled (false) or disabled (true).
	Disabled bool `json:"disabled,omitempty"`

	// Listen port for scraping metrics.
	ListenPort int64 `json:"listen_port,omitempty"`

	// PostgreSQL password for scraping metrics.
	Password string `json:"password,omitempty"`

	// The pmm-agent identifier which runs this instance.
	PMMAgentID string `json:"pmm_agent_id,omitempty"`

	// Service identifier.
	ServiceID string `json:"service_id,omitempty"`

	// AgentStatus represents actual Agent status.
	// Enum: [AGENT_STATUS_INVALID STARTING RUNNING WAITING STOPPING DONE]
	Status *string `json:"status,omitempty"`

	// PostgreSQL username for scraping metrics.
	Username string `json:"username,omitempty"`
}

// Validate validates this add postgre SQL OK body postgres exporter
func (o *AddPostgreSQLOKBodyPostgresExporter) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.validateStatus(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var addPostgreSqlOkBodyPostgresExporterTypeStatusPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["AGENT_STATUS_INVALID","STARTING","RUNNING","WAITING","STOPPING","DONE"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		addPostgreSqlOkBodyPostgresExporterTypeStatusPropEnum = append(addPostgreSqlOkBodyPostgresExporterTypeStatusPropEnum, v)
	}
}

const (

	// AddPostgreSQLOKBodyPostgresExporterStatusAGENTSTATUSINVALID captures enum value "AGENT_STATUS_INVALID"
	AddPostgreSQLOKBodyPostgresExporterStatusAGENTSTATUSINVALID string = "AGENT_STATUS_INVALID"

	// AddPostgreSQLOKBodyPostgresExporterStatusSTARTING captures enum value "STARTING"
	AddPostgreSQLOKBodyPostgresExporterStatusSTARTING string = "STARTING"

	// AddPostgreSQLOKBodyPostgresExporterStatusRUNNING captures enum value "RUNNING"
	AddPostgreSQLOKBodyPostgresExporterStatusRUNNING string = "RUNNING"

	// AddPostgreSQLOKBodyPostgresExporterStatusWAITING captures enum value "WAITING"
	AddPostgreSQLOKBodyPostgresExporterStatusWAITING string = "WAITING"

	// AddPostgreSQLOKBodyPostgresExporterStatusSTOPPING captures enum value "STOPPING"
	AddPostgreSQLOKBodyPostgresExporterStatusSTOPPING string = "STOPPING"

	// AddPostgreSQLOKBodyPostgresExporterStatusDONE captures enum value "DONE"
	AddPostgreSQLOKBodyPostgresExporterStatusDONE string = "DONE"
)

// prop value enum
func (o *AddPostgreSQLOKBodyPostgresExporter) validateStatusEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, addPostgreSqlOkBodyPostgresExporterTypeStatusPropEnum); err != nil {
		return err
	}
	return nil
}

func (o *AddPostgreSQLOKBodyPostgresExporter) validateStatus(formats strfmt.Registry) error {

	if swag.IsZero(o.Status) { // not required
		return nil
	}

	// value enum
	if err := o.validateStatusEnum("addPostgreSqlOk"+"."+"postgres_exporter"+"."+"status", "body", *o.Status); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (o *AddPostgreSQLOKBodyPostgresExporter) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *AddPostgreSQLOKBodyPostgresExporter) UnmarshalBinary(b []byte) error {
	var res AddPostgreSQLOKBodyPostgresExporter
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

/*AddPostgreSQLOKBodyService PostgreSQLService represents a generic PostgreSQL instance.
swagger:model AddPostgreSQLOKBodyService
*/
type AddPostgreSQLOKBodyService struct {

	// Access address (DNS name or IP).
	Address string `json:"address,omitempty"`

	// Custom user-assigned labels.
	CustomLabels map[string]string `json:"custom_labels,omitempty"`

	// Environment name.
	Environment string `json:"environment,omitempty"`

	// Node identifier where this instance runs.
	NodeID string `json:"node_id,omitempty"`

	// Access port.
	Port int64 `json:"port,omitempty"`

	// Unique randomly generated instance identifier.
	ServiceID string `json:"service_id,omitempty"`

	// Unique across all Services user-defined name.
	ServiceName string `json:"service_name,omitempty"`
}

// Validate validates this add postgre SQL OK body service
func (o *AddPostgreSQLOKBodyService) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *AddPostgreSQLOKBodyService) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *AddPostgreSQLOKBodyService) UnmarshalBinary(b []byte) error {
	var res AddPostgreSQLOKBodyService
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
