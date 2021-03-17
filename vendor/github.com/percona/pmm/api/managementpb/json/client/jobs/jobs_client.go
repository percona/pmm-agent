// Code generated by go-swagger; DO NOT EDIT.

package jobs

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new jobs API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for jobs API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientService is the interface for Client methods
type ClientService interface {
	CancelActionMixin5(params *CancelActionMixin5Params) (*CancelActionMixin5OK, error)

	GetActionMixin5(params *GetActionMixin5Params) (*GetActionMixin5OK, error)

	StartEchoJob(params *StartEchoJobParams) (*StartEchoJobOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
  CancelActionMixin5 cancels action stops a job
*/
func (a *Client) CancelActionMixin5(params *CancelActionMixin5Params) (*CancelActionMixin5OK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCancelActionMixin5Params()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "CancelActionMixin5",
		Method:             "POST",
		PathPattern:        "/v1/management/Jobs/Cancel",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &CancelActionMixin5Reader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CancelActionMixin5OK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*CancelActionMixin5Default)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetActionMixin5 gets action gets an result of given action
*/
func (a *Client) GetActionMixin5(params *GetActionMixin5Params) (*GetActionMixin5OK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetActionMixin5Params()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "GetActionMixin5",
		Method:             "POST",
		PathPattern:        "/v1/management/Jobs/Get",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetActionMixin5Reader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetActionMixin5OK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetActionMixin5Default)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  StartEchoJob starts echo job starts echo job
*/
func (a *Client) StartEchoJob(params *StartEchoJobParams) (*StartEchoJobOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewStartEchoJobParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "StartEchoJob",
		Method:             "POST",
		PathPattern:        "/v1/management/Jobs/StartEcho",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &StartEchoJobReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*StartEchoJobOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*StartEchoJobDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
