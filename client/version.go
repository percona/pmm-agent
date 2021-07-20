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

// Package client contains business logic of working with pmm-managed.
package client

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/percona/pmm/api/agentpb"
)


func (c *Client) handleVersionRequest(r *agentpb.GetVersionRequest) (string, *status.Status) {
	var version string
	var err error
	switch r.Software.(type) {
	case *agentpb.GetVersionRequest_LocalMysql:
		version, err = c.localMySQLVersion()
	case *agentpb.GetVersionRequest_Xtrabackup:
		version, err = c.xtrabackupVersion()
	default:
		return "", status.Newf(codes.Unknown, "unknown software type %v.", r)
	}

	if err != nil {
		return "", status.New(codes.Internal, err.Error())
	}

	return version, nil
}
