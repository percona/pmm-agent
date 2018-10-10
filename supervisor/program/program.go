package program

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/percona/pmm-agent/errs"
)

// Program contains information necessary to execute program.
type Program struct {
	// Program name, should be unique to avoid conflict with other programs.
	Program string
	// Name is executable to run.
	Name string
	// Arg is a list of arguments passed to executable.
	Arg []string
	// Env is a list of environment variables.
	Env []string

	cmd     *exec.Cmd
	outfile *os.File
	errfile *os.File
	cancel  context.CancelFunc
	done    chan struct{}
	err     error
	sync.RWMutex
}

// Start program.
func (p *Program) Start() error {
	p.Lock()
	defer p.Unlock()

	if p.running() {
		return fmt.Errorf("program '%s' is already running", p.Program)
	}

	return p.run()
}

// Stop program.
func (p *Program) Stop() error {
	cancel, done, err := p.cancelAndDone()
	if err != nil {
		return err
	}
	cancel()
	<-done

	// "signal: killed" is expected as cancel() kills process.
	// And ugly hack for now to filter out the error.
	err = p.Err()
	if err.Error() == "signal: killed" {
		return nil
	}

	return err
}

// Running returns true if program is running.
func (p *Program) Running() bool {
	p.RLock()
	defer p.RUnlock()
	return p.running()
}

// Err returns error if program quit in non standard way.
func (p *Program) Err() error {
	p.RLock()
	defer p.RUnlock()
	return p.err
}

// CombinedOutput of program.
func (p *Program) CombinedOutput() []byte {
	stdout, stderr := p.stdOutErr()
	return combinedOutput(stdout, stderr)
}

// Pid of running process.
func (p *Program) Pid() int {
	p.RLock()
	defer p.RUnlock()
	if !p.running() {
		return 0
	}
	return p.cmd.Process.Pid
}

func (p *Program) running() bool {
	if p.done == nil {
		return false
	}
	select {
	default:
	case <-p.done:
		return false
	}
	return true
}

func (p *Program) run() (err error) {
	f, err := os.Create(filepath.Join(fmt.Sprintf("pmm-%s.log", p.Name)))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, p.Name, p.Arg...)
	cmd.Env = p.Env
	cmd.Stderr = f
	cmd.Stdout = f
	if err := cmd.Start(); err != nil {
		cancel()
		return err
	}

	p.outfile = f
	p.errfile = f
	p.cmd = cmd
	p.err = nil
	p.cancel = cancel
	p.done = make(chan struct{})
	go p.wait()

	// Wait until it starts to report log.
	for {
		select {
		case <-time.After(100 * time.Millisecond):
			if combinedOutput(f) != nil {
				return nil
			}
		case <-time.After(1 * time.Second):
			// Do not wait anymore.
			return nil
		}
	}

	return nil
}

func (p *Program) wait() {
	var errs errs.Errs

	if err := p.cmd.Wait(); err != nil {
		errs = append(errs, err)
	}

	p.Lock()
	defer p.Unlock()

	if err := p.outfile.Close(); err != nil {
		errs = append(errs, err)
	}
	if p.outfile != p.errfile {
		if err := p.errfile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	p.cmd = nil
	p.cancel = nil
	switch len(errs) {
	case 0:
		p.err = nil
	default:
		p.err = errs
	}
	close(p.done)
}

func (p *Program) cancelAndDone() (context.CancelFunc, chan struct{}, error) {
	p.RLock()
	defer p.RUnlock()
	if !p.running() {
		return nil, nil, fmt.Errorf("program '%s' is already stopped", p.Program)
	}
	return p.cancel, p.done, nil
}

func (p *Program) stdOutErr() (*os.File, *os.File) {
	p.RLock()
	defer p.RUnlock()
	return p.outfile, p.errfile
}

func combinedOutput(files ...*os.File) (out []byte) {
	var seen = map[*os.File]struct{}{}
	for i := range files {
		if _, ok := seen[files[i]]; ok {
			continue
		}
		data, err := ioutil.ReadFile(files[i].Name())
		if err != nil {
			out = append(out, []byte(err.Error())...)
		}
		out = append(out, data...)
		seen[files[i]] = struct{}{}
	}
	return
}
