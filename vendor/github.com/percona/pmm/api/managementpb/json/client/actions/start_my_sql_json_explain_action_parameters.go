// Code generated by go-swagger; DO NOT EDIT.

package actions

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"
)

// NewStartMySQLJSONExplainActionParams creates a new StartMySQLJSONExplainActionParams object
// with the default values initialized.
func NewStartMySQLJSONExplainActionParams() *StartMySQLJSONExplainActionParams {
	var ()
	return &StartMySQLJSONExplainActionParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewStartMySQLJSONExplainActionParamsWithTimeout creates a new StartMySQLJSONExplainActionParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewStartMySQLJSONExplainActionParamsWithTimeout(timeout time.Duration) *StartMySQLJSONExplainActionParams {
	var ()
	return &StartMySQLJSONExplainActionParams{

		timeout: timeout,
	}
}

// NewStartMySQLJSONExplainActionParamsWithContext creates a new StartMySQLJSONExplainActionParams object
// with the default values initialized, and the ability to set a context for a request
func NewStartMySQLJSONExplainActionParamsWithContext(ctx context.Context) *StartMySQLJSONExplainActionParams {
	var ()
	return &StartMySQLJSONExplainActionParams{

		Context: ctx,
	}
}

// NewStartMySQLJSONExplainActionParamsWithHTTPClient creates a new StartMySQLJSONExplainActionParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewStartMySQLJSONExplainActionParamsWithHTTPClient(client *http.Client) *StartMySQLJSONExplainActionParams {
	var ()
	return &StartMySQLJSONExplainActionParams{
		HTTPClient: client,
	}
}

/*StartMySQLJSONExplainActionParams contains all the parameters to send to the API endpoint
for the start my SQL Json explain action operation typically these are written to a http.Request
*/
type StartMySQLJSONExplainActionParams struct {

	/*Body*/
	Body StartMySQLJSONExplainActionBody

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the start my SQL Json explain action params
func (o *StartMySQLJSONExplainActionParams) WithTimeout(timeout time.Duration) *StartMySQLJSONExplainActionParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the start my SQL Json explain action params
func (o *StartMySQLJSONExplainActionParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the start my SQL Json explain action params
func (o *StartMySQLJSONExplainActionParams) WithContext(ctx context.Context) *StartMySQLJSONExplainActionParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the start my SQL Json explain action params
func (o *StartMySQLJSONExplainActionParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the start my SQL Json explain action params
func (o *StartMySQLJSONExplainActionParams) WithHTTPClient(client *http.Client) *StartMySQLJSONExplainActionParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the start my SQL Json explain action params
func (o *StartMySQLJSONExplainActionParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the start my SQL Json explain action params
func (o *StartMySQLJSONExplainActionParams) WithBody(body StartMySQLJSONExplainActionBody) *StartMySQLJSONExplainActionParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the start my SQL Json explain action params
func (o *StartMySQLJSONExplainActionParams) SetBody(body StartMySQLJSONExplainActionBody) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *StartMySQLJSONExplainActionParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if err := r.SetBodyParam(o.Body); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
