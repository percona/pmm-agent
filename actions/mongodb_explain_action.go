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
	"strings"

	"github.com/percona/percona-toolkit/src/go/mongolib/explain"
	"github.com/percona/pmm/api/agentpb"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongodbExplainAction struct {
	id     string
	params *agentpb.StartActionRequest_MongoDBExplainParams
}

// NewMongoDBExplain creates a MongoDB  EXPLAIN query Action.
func NewMongoDBExplain(id string, params *agentpb.StartActionRequest_MongoDBExplainParams) Action {
	return &mongodbExplainAction{
		id:     id,
		params: params,
	}
}

// ID returns an Action ID.
func (a *mongodbExplainAction) ID() string {
	return a.id
}

// Type returns an Action type.
func (a *mongodbExplainAction) Type() string {
	return "mongodb-explain"
}

// Run runs an Action and returns output and error.
func (a *mongodbExplainAction) Run(ctx context.Context) ([]byte, error) {
	dsn := a.params.Dsn
	if !strings.HasPrefix(dsn, "mongodb://") {
		dsn = "mongodb://" + dsn
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dsn))

	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot connect to the database. Ping failed")
	}
	defer client.Disconnect(ctx) //nolint:errcheck

	ex := explain.New(ctx, client)
	return ex.Run(a.params.Database, []byte(a.params.Query))
}

func (a *mongodbExplainAction) sealed() {}
