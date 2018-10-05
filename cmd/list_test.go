package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/percona/pmm-agent/api"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := &Server{
		Addr: "127.0.0.1:7771",
	}
	go func() {
		err := server.Serve(ctx)
		assert.NoError(t, err)
	}()

	var buf *bytes.Buffer

	// Initial list should be empty.
	{
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
		for i := range resp.Statuses {
			assert.NotEmpty(t, resp.Statuses[i].Pid)
			resp.Statuses[i].Pid = 0
		}
		assert.Equal(t, expected, resp)
	}

	// Stop program.
	{
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
		// PID is dynamic so we can't test it but we can ensure it's not empty.
		for i := range resp.Statuses {
			assert.Empty(t, resp.Statuses[i].Pid)
			resp.Statuses[i].Pid = 0
		}
		assert.Equal(t, expected, resp)
	}

	// Remove all programs.
	{
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
		rootCmd.SetArgs([]string{
			"list",
		})
		buf = &bytes.Buffer{}
		rootCmd.SetOutput(buf)
		assert.NoError(t, rootCmd.Execute())
		assert.Equal(t, "", buf.String())
	}
}
