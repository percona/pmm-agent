// Code generated by go-swagger; DO NOT EDIT.

package actions

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"
)

// New creates a new actions API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) *Client {
	return &Client{transport: transport, formats: formats}
}

/*
Client for actions API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

/*
CancelAction cancels action stops an action
*/
func (a *Client) CancelAction(params *CancelActionParams) (*CancelActionOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCancelActionParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "CancelAction",
		Method:             "POST",
		PathPattern:        "/v1/management/Actions/CancelAction",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &CancelActionReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*CancelActionOK), nil

}

/*
RunAction runs action runs an action
*/
func (a *Client) RunAction(params *RunActionParams) (*RunActionOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewRunActionParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "RunAction",
		Method:             "POST",
		PathPattern:        "/v1/management/Actions/RunAction",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &RunActionReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*RunActionOK), nil

}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
