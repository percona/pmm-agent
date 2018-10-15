package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/percona/pmm-agent/api"
	"github.com/percona/pmm-agent/app"
	"github.com/percona/pmm-agent/app/server"
)

func TestList(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		app := &app.App{
			Server: server.Server{
				Addr: "127.0.0.1:7771",
			},
		}

		err := app.Server.Serve(ctx)
		assert.NoError(t, err)
	}()

	var buf *bytes.Buffer

	// Initial list should be empty.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"list",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		assert.Equal(t, "", buf.String())
	}

	// Add new program.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"add", "mysql-1", "--env", "DATA_SOURCE_NAME=root@/", "--", "mysqld_exporter", "--collect.all",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		assert.Equal(t, "", buf.String())
	}

	// List now should contain new program.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"list", "--json",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		resp := &api.ListResponse{}
		err := json.Unmarshal(buf.Bytes(), &resp)
		assert.NoError(t, err)
		expected := &api.ListResponse{
			Statuses: map[string]*api.Status{
				"mysql-1": {
					Program: &api.Program{
						Name: "mysqld_exporter",
						Arg: []string{
							"--collect.all",
						},
						Env: []string{
							"DATA_SOURCE_NAME=root@/",
						},
					},
					Running: true,
				},
			},
		}
		// PID is dynamic so we can't test it but we can ensure it's not empty.
		// OUT is dynamic so we can't test it but we can ensure it's not empty.
		for i := range resp.Statuses {
			assert.NotEmpty(t, resp.Statuses[i].Pid)
			resp.Statuses[i].Pid = 0
			assert.NotEmpty(t, resp.Statuses[i].Out)
			resp.Statuses[i].Out = ""
		}
		assert.Equal(t, expected, resp)
	}

	// Stop program.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"stop", "mysql-1",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		assert.Equal(t, "", buf.String())
	}

	// List now should contain stopped program.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"list", "--json",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		resp := &api.ListResponse{}
		err := json.Unmarshal(buf.Bytes(), &resp)
		assert.NoError(t, err)
		expected := &api.ListResponse{
			Statuses: map[string]*api.Status{
				"mysql-1": {
					Program: &api.Program{
						Name: "mysqld_exporter",
						Arg: []string{
							"--collect.all",
						},
						Env: []string{
							"DATA_SOURCE_NAME=root@/",
						},
					},
					Running: false,
					Err:     "signal: killed",
				},
			},
		}
		// PID is dynamic so we can't test it but we can ensure it's empty.
		// OUT is dynamic so we can't test it but we can ensure it's not empty.
		for i := range resp.Statuses {
			assert.Empty(t, resp.Statuses[i].Pid)
			resp.Statuses[i].Pid = 0
			assert.NotEmpty(t, resp.Statuses[i].Out)
			resp.Statuses[i].Out = ""
		}
		assert.Equal(t, expected, resp)
	}

	// Start program.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"start", "mysql-1",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		assert.Equal(t, "", buf.String())
	}

	// List now should contain started program.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"list", "--json",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		resp := &api.ListResponse{}
		err := json.Unmarshal(buf.Bytes(), &resp)
		assert.NoError(t, err)
		expected := &api.ListResponse{
			Statuses: map[string]*api.Status{
				"mysql-1": {
					Program: &api.Program{
						Name: "mysqld_exporter",
						Arg: []string{
							"--collect.all",
						},
						Env: []string{
							"DATA_SOURCE_NAME=root@/",
						},
					},
					Running: true,
				},
			},
		}
		// PID is dynamic so we can't test it but we can ensure it's empty.
		// OUT is dynamic so we can't test it but we can ensure it's not empty.
		for i := range resp.Statuses {
			assert.NotEmpty(t, resp.Statuses[i].Pid)
			resp.Statuses[i].Pid = 0
			assert.NotEmpty(t, resp.Statuses[i].Out)
			resp.Statuses[i].Out = ""
		}
		assert.Equal(t, expected, resp)
	}

	// Add another new program.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"add", "mysql-2", "--env", "DATA_SOURCE_NAME=root@/", "--",
			"mysqld_exporter",
			"--collect.all",
			"--web.listen-address", ":9204",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		assert.Equal(t, "", buf.String())
	}

	// List now should contain started programs.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"list", "--json",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		resp := &api.ListResponse{}
		err := json.Unmarshal(buf.Bytes(), &resp)
		assert.NoError(t, err)
		expected := &api.ListResponse{
			Statuses: map[string]*api.Status{
				"mysql-1": {
					Program: &api.Program{
						Name: "mysqld_exporter",
						Arg: []string{
							"--collect.all",
						},
						Env: []string{
							"DATA_SOURCE_NAME=root@/",
						},
					},
					Running: true,
				},
				"mysql-2": {
					Program: &api.Program{
						Name: "mysqld_exporter",
						Arg: []string{
							"--collect.all",
							"--web.listen-address", ":9204",
						},
						Env: []string{
							"DATA_SOURCE_NAME=root@/",
						},
					},
					Running: true,
				},
			},
		}
		// PID is dynamic so we can't test it but we can ensure it's empty.
		// OUT is dynamic so we can't test it but we can ensure it's not empty.
		for i := range resp.Statuses {
			assert.NotEmpty(t, resp.Statuses[i].Pid)
			resp.Statuses[i].Pid = 0
			assert.NotEmpty(t, resp.Statuses[i].Out)
			resp.Statuses[i].Out = ""
		}
		assert.Equal(t, expected, resp)
	}

	// Stop programs.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"stop", "mysql-1", "mysql-2",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		assert.Equal(t, "", buf.String())
	}

	// Start programs.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"start", "mysql-1", "mysql-2",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		assert.Equal(t, "", buf.String())
	}

	// Stop all programs.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"stop",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		assert.Equal(t, "", buf.String())
	}

	// Start all programs.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"start",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		assert.Equal(t, "", buf.String())
	}

	// Remove program.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"remove", "mysql-1",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		assert.Equal(t, "", buf.String())
	}

	// Remove all programs.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"remove",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		assert.Equal(t, "", buf.String())
	}

	// List should be empty again.
	{
		rootCmd := New(&app.App{})
		rootCmd.SetArgs([]string{
			"list",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		assert.Equal(t, "", buf.String())
	}
}
