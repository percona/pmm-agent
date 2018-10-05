package errs

import "sync"

// Safe is a list of errors safe for concurrent use by multiple goroutines.
type Safe struct {
	errs Errs
	sync.RWMutex
}

// Err returns error.
func (s *Safe) Err() error {
	s.RLock()
	defer s.RUnlock()
	if len(s.errs) > 0 {
		return s.errs
	}
	return nil
}

// Add error.
func (s *Safe) Add(err error) {
	if err == nil {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.errs = append(s.errs, err)
}

// Error formats error to string.
func (s *Safe) Error() string {
	s.RLock()
	defer s.RUnlock()
	return s.errs.Error()
}
