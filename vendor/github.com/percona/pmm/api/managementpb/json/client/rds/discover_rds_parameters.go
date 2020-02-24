// Code generated by go-swagger; DO NOT EDIT.

package rds

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

// NewDiscoverRDSParams creates a new DiscoverRDSParams object
// with the default values initialized.
func NewDiscoverRDSParams() *DiscoverRDSParams {
	var ()
	return &DiscoverRDSParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewDiscoverRDSParamsWithTimeout creates a new DiscoverRDSParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewDiscoverRDSParamsWithTimeout(timeout time.Duration) *DiscoverRDSParams {
	var ()
	return &DiscoverRDSParams{

		timeout: timeout,
	}
}

// NewDiscoverRDSParamsWithContext creates a new DiscoverRDSParams object
// with the default values initialized, and the ability to set a context for a request
func NewDiscoverRDSParamsWithContext(ctx context.Context) *DiscoverRDSParams {
	var ()
	return &DiscoverRDSParams{

		Context: ctx,
	}
}

// NewDiscoverRDSParamsWithHTTPClient creates a new DiscoverRDSParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewDiscoverRDSParamsWithHTTPClient(client *http.Client) *DiscoverRDSParams {
	var ()
	return &DiscoverRDSParams{
		HTTPClient: client,
	}
}

/*DiscoverRDSParams contains all the parameters to send to the API endpoint
for the discover RDS operation typically these are written to a http.Request
*/
type DiscoverRDSParams struct {

	/*Body*/
	Body DiscoverRDSBody

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the discover RDS params
func (o *DiscoverRDSParams) WithTimeout(timeout time.Duration) *DiscoverRDSParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the discover RDS params
func (o *DiscoverRDSParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the discover RDS params
func (o *DiscoverRDSParams) WithContext(ctx context.Context) *DiscoverRDSParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the discover RDS params
func (o *DiscoverRDSParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the discover RDS params
func (o *DiscoverRDSParams) WithHTTPClient(client *http.Client) *DiscoverRDSParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the discover RDS params
func (o *DiscoverRDSParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the discover RDS params
func (o *DiscoverRDSParams) WithBody(body DiscoverRDSBody) *DiscoverRDSParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the discover RDS params
func (o *DiscoverRDSParams) SetBody(body DiscoverRDSBody) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *DiscoverRDSParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
