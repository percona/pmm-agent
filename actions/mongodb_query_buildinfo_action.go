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

package actions

import (
	"context"

	"github.com/percona/pmm/api/agentpb"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongodbQueryBuildinfoAction struct {
	id     string
	params *agentpb.StartActionRequest_MongoDBQueryBuildInfoParams
}

// NewMongoDBQueryBuildinfoAction creates a MongoDB Buildinfo query Action.
func NewMongoDBQueryBuildinfoAction(id string, params *agentpb.StartActionRequest_MongoDBQueryBuildInfoParams) Action {
	return &mongodbQueryBuildinfoAction{
		id:     id,
		params: params,
	}
}

// ID returns an Action ID.
func (a *mongodbQueryBuildinfoAction) ID() string {
	return a.id
}

// Type returns an Action type.
func (a *mongodbQueryBuildinfoAction) Type() string {
	return "mongodb-query-buildinfo"
}

// Run runs an Action and returns output and error.
func (a *mongodbQueryBuildinfoAction) Run(ctx context.Context) ([]byte, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(a.params.Dsn))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer client.Disconnect(ctx) //nolint:errcheck

	runCommand := bson.D{{"buildInfo", 1}} //nolint:govet
	res := client.Database("admin").RunCommand(ctx, runCommand)

	var doc map[string]interface{}
	if err = res.Decode(&doc); err != nil {
		return nil, errors.WithStack(err)
	}

	data := []map[string]interface{}{doc}
	return agentpb.MarshalActionQueryDocsResult(data)
}

func (a *mongodbQueryBuildinfoAction) sealed() {}
