package supervisor

import (
	"fmt"
	"sync"

	"github.com/percona/pmm-agent/errs"

	"github.com/percona/pmm-agent/supervisor/program"
)

// Supervisor allows to manage programs.
type Supervisor struct {
	LogDir   string
	programs sync.Map
}

// List programs.
func (s *Supervisor) List() (programs []*program.Program, err error) {
	s.programs.Range(func(key, value interface{}) bool {
		p, ok := value.(*program.Program)
		if !ok {
			err = fmt.Errorf("internal error: value doesn't look like program: %s", key)
			return false
		}
		programs = append(programs, p)
		return true
	})

	return programs, err
}

// Add new program with given name and options.
func (s *Supervisor) Add(name string, p *program.Program) error {
	_, ok := s.programs.Load(name)
	if ok {
		return fmt.Errorf("program already exists: %s", name)
	}

	err := p.Start()
	if err != nil {
		return err
	}
	s.programs.Store(name, p)

	return p.Err()
}

// Remove program by name.
func (s *Supervisor) Remove(name string) error {
	if err := s.Stop(name); err != nil {
		return err
	}

	s.programs.Delete(name)
	return nil
}

// Start program by name.
func (s *Supervisor) Start(name string) error {
	p, err := s.program(name)
	if err != nil {
		return err
	}
	return p.Start()
}

// Stop program by name.
func (s *Supervisor) Stop(name string) error {
	p, err := s.program(name)
	if err != nil {
		return err
	}

	if p.Running() {
		return p.Stop()
	}

	return nil
}

// StopAll programs.
func (s *Supervisor) StopAll() error {
	var errs errs.Errs
	s.programs.Range(func(key, value interface{}) bool {
		p, ok := value.(*program.Program)
		if !ok {
			errs = append(errs, fmt.Errorf("internal error: value doesn't look like program: %s", key))
			return true
		}
		if p.Running() {
			if err := p.Stop(); err != nil {
				errs = append(errs, err)
			}
		}
		return true
	})
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// StartAll programs.
func (s *Supervisor) StartAll() error {
	var errs errs.Errs
	s.programs.Range(func(key, value interface{}) bool {
		p, ok := value.(*program.Program)
		if !ok {
			errs = append(errs, fmt.Errorf("internal error: value doesn't look like program: %s", key))
			return true
		}
		if !p.Running() {
			if err := p.Start(); err != nil {
				errs = append(errs, err)
			}
		}
		return true
	})
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// RemoveAll programs.
func (s *Supervisor) RemoveAll() error {
	var errs errs.Errs
	s.programs.Range(func(key, value interface{}) bool {
		p, ok := value.(*program.Program)
		if !ok {
			errs = append(errs, fmt.Errorf("internal error: value doesn't look like program: %s", key))
			return true
		}
		if p.Running() {
			if err := p.Stop(); err != nil {
				errs = append(errs, err)
			}
		}
		s.programs.Delete(key)
		return true
	})
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// program returns program identified by name or error otherwise.
func (s *Supervisor) program(name string) (*program.Program, error) {
	v, ok := s.programs.Load(name)
	if !ok {
		return nil, fmt.Errorf("program doesn't exists: %s", name)
	}
	p, ok := v.(*program.Program)
	if !ok {
		return nil, fmt.Errorf("internal error: value doesn't look like program: %s", name)
	}

	return p, nil
}
