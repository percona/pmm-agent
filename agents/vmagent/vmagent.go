// pmm-agent
// Copyright 2019 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vmagent

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime/pprof"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/percona/pmm-agent/agents/process"
	"github.com/percona/pmm-agent/config"
)

// VMAgent represents VictoriaMetrics agent.
type VMAgent struct {
	remoteInsecure   bool
	remoteUserName   string
	remotePassword   string
	binaryPath       string
	client           *http.Client
	remoteWriteURL   *url.URL
	listenURL        *url.URL
	scrapeConfigPath string
	lastCfg          []byte
	tmpDir           string
	l                *logrus.Entry
	cancel           context.CancelFunc
	done             chan struct{}
}

// NewVMAgent - creates vmagent object and scrape config file.
func NewVMAgent(cfg *config.Config) (*VMAgent, error) {
	remoteWriteURL := *cfg.Server.URL()
	remoteWriteURL.Path = path.Join(remoteWriteURL.Path, "victoriametrics", "api", "v1", "write")
	remoteWriteURL.User = nil
	scrapeCfgPath := path.Join(cfg.Paths.TempDir, "vmagent-scrape-config.yaml")
	scrapeCfg, err := ioutil.ReadFile(scrapeCfgPath) //nolint:gosec
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "failed read vmagent-scrape-config file at %q", scrapeCfgPath)
		}
		f, err := os.Create(scrapeCfgPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create vmagent-scrape-config file at %q", scrapeCfgPath)
		}
		_ = f.Close() //nolint:errcheck
	}

	return &VMAgent{
		remoteInsecure: cfg.Server.InsecureTLS,
		remoteUserName: cfg.Server.Username,
		remotePassword: cfg.Server.Password,
		binaryPath:     cfg.Paths.VMAgent,
		client:         new(http.Client),
		remoteWriteURL: &remoteWriteURL,
		listenURL: &url.URL{
			Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(int(cfg.Ports.VMAgent))),
			Scheme: "http",
		},
		scrapeConfigPath: scrapeCfgPath,
		lastCfg:          scrapeCfg,
		tmpDir:           path.Join(cfg.Paths.TempDir, "vmagent-tmp-dir"),
		l: logrus.WithFields(logrus.Fields{
			"component": "agent-process",
			"agentID":   "vmagent",
			"type":      "vmagent",
		}),
	}, nil
}

// Start starts vmagent process.
func (vma *VMAgent) Start(ctx context.Context, cfgChan chan []byte) {
	ctx, cancel := context.WithCancel(ctx)
	pr := &process.Params{Path: vma.binaryPath, Args: vma.args()}
	vma.l.Debugf("Starting: %s.", pr)
	process := process.New(pr, nil, vma.l)
	go pprof.Do(ctx, pprof.Labels("agentID", "vmagent", "type", "vmagent"), process.Run)
	done := make(chan struct{})
	go func() {
		for status := range process.Changes() {
			vma.l.Infof("vmanget status changed: %s ", status)
		}
		close(done)
	}()
	go vma.listenForCfgUpdates(ctx, cfgChan)
	vma.cancel = cancel
	vma.done = done
}

// args returns vmagent process args.
func (vma *VMAgent) args() []string {
	baseArgs := []string{
		fmt.Sprintf("-remoteWrite.url=%s", vma.remoteWriteURL.String()),
		fmt.Sprintf("-remoteWrite.basicAuth.username=%s", vma.remoteUserName),
		fmt.Sprintf("-remoteWrite.basicAuth.password=%s", vma.remotePassword),
		fmt.Sprintf("-remoteWrite.tmpDataPath=%s", vma.tmpDir),
		fmt.Sprintf("-promscrape.config=%s", vma.scrapeConfigPath),
		// 1GB disk queue size
		"-remoteWrite.maxDiskUsagePerURL=1073741824",
		// reduces log verbose
		"-loggerLevel=WARN",
		fmt.Sprintf("-httpListenAddr=%s", vma.listenURL.Host),
	}
	if vma.remoteInsecure {
		baseArgs = append(baseArgs, "-remoteWrite.tlsInsecureSkipVerify=true")
	}

	return baseArgs
}

// listenForCfgUpdates listens for cfg updates and triggers config update.
func (vma *VMAgent) listenForCfgUpdates(ctx context.Context, newCfg chan []byte) {
	for {
		select {
		case <-ctx.Done():
			vma.l.Infof("stopping agent")
			return
		case cfg := <-newCfg:
			vma.updateScrapeConfig(cfg)
		}
	}
}

// updateScrapeConfig writes new scrape config file and triggers config file re-read.
func (vma *VMAgent) updateScrapeConfig(data []byte) {
	if bytes.Equal(data, vma.lastCfg) {
		return
	}
	err := ioutil.WriteFile(vma.scrapeConfigPath, data, 0600)
	if err != nil {
		vma.l.WithError(err).Errorf("cannot write scrape config to: %q", vma.scrapeConfigPath)
		return
	}
	u := *vma.listenURL
	u.Path = path.Join(u.Path, "-", "reload")
	req, err := http.NewRequestWithContext(context.Background(), "GET", u.String(), nil)
	if err != nil {
		vma.l.WithError(err).Error("cannot create vmagent reload request")
		return
	}
	resp, err := vma.client.Do(req)
	if err != nil {
		vma.l.WithError(err).Errorf("failed query vmagent reload api")
		return
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		vma.l.Errorf("unexpected status code: %d , want: %d", resp.StatusCode, http.StatusOK)
		return
	}
	vma.lastCfg = data
	vma.l.Info("successfully reloaded vmagent config")
}

// Stop shutdowns vmagent.
func (vma *VMAgent) Stop() {
	vma.cancel()
	<-vma.done
}
