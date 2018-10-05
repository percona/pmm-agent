package handlers

import (
	"golang.org/x/net/context"

	"github.com/percona/pmm-agent/api"
	"github.com/percona/pmm-agent/supervisor"
	"github.com/percona/pmm-agent/supervisor/program"
)

// Ensure struct matches interface.
var _ api.SupervisorServer = (*SupervisorServer)(nil)

// SupervisorServer allows to manage programs.
type SupervisorServer struct {
	Supervisor *supervisor.Supervisor
}

// List all programs.
func (s *SupervisorServer) List(ctx context.Context, req *api.ListRequest) (*api.ListResponse, error) {
	ps, err := s.Supervisor.List()
	if err != nil {
		return nil, err
	}
	statuses := make(map[string]*api.Status, len(ps))
	for _, p := range ps {
		status := &api.Status{
			Program: &api.Program{
				Name: p.Name,
				Arg:  p.Arg,
				Env:  p.Env,
			},
			Pid:     int64(p.Pid()),
			Running: p.Running(),
			Out:     string(p.CombinedOutput()),
		}
		err := p.Err()
		if err != nil {
			status.Err = err.Error()
		}
		statuses[p.Program] = status
	}
	resp := &api.ListResponse{
		Statuses: statuses,
	}
	return resp, nil
}

// Add program.
func (s *SupervisorServer) Add(ctx context.Context, req *api.AddRequest) (*api.AddResponse, error) {
	programName := req.Name
	p := &program.Program{
		Program: programName,
		Name:    req.Program.Name,
		Arg:     req.Program.Arg,
		Env:     req.Program.Env,
	}
	resp := &api.AddResponse{}
	err := s.Supervisor.Add(programName, p)
	return resp, err
}

// Remove program.
func (s *SupervisorServer) Remove(ctx context.Context, req *api.RemoveRequest) (*api.RemoveResponse, error) {
	resp := &api.RemoveResponse{}
	err := s.Supervisor.Remove(req.Name)
	return resp, err
}

// Remove all programs.
func (s *SupervisorServer) RemoveAll(ctx context.Context, req *api.RemoveAllRequest) (*api.RemoveAllResponse, error) {
	return &api.RemoveAllResponse{}, s.Supervisor.RemoveAll()
}

// Start program.
func (s *SupervisorServer) Start(ctx context.Context, req *api.StartRequest) (*api.StartResponse, error) {
	err := s.Supervisor.Start(req.Name)
	if err != nil {
		return nil, err
	}
	resp := &api.StartResponse{}
	return resp, nil
}

// StartAll programs.
func (s *SupervisorServer) StartAll(ctx context.Context, req *api.StartAllRequest) (*api.StartAllResponse, error) {
	return &api.StartAllResponse{}, s.Supervisor.StartAll()
}

// Stop program.
func (s *SupervisorServer) Stop(ctx context.Context, req *api.StopRequest) (*api.StopResponse, error) {
	err := s.Supervisor.Stop(req.Name)
	if err != nil {
		return nil, err
	}
	resp := &api.StopResponse{}
	return resp, nil
}

// StopAll programs.
func (s *SupervisorServer) StopAll(ctx context.Context, req *api.StopAllRequest) (*api.StopAllResponse, error) {
	return &api.StopAllResponse{}, s.Supervisor.StopAll()
}
