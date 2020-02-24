// pmm-agent
// Copyright (C) 2018 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package actions

import (
	"context"
	"fmt"

	"github.com/percona/percona-toolkit/src/go/mongolib/proto"
	"github.com/percona/pmm/api/agentpb"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
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
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot connect to the database. Ping failed")
	}
	defer client.Disconnect(ctx) //nolint:errcheck

	var eq proto.ExampleQuery

	err = bson.UnmarshalExtJSON([]byte(a.params.Query), true, &eq)
	if err != nil {
		return nil, fmt.Errorf("explain: unable to decode query %s: %s", a.params.Query, err)
	}

	var result bson.D
	res := client.Database(eq.Db()).RunCommand(ctx, eq.ExplainCmd())
	if res.Err() != nil {
		return nil, res.Err()
	}

	if err := res.Decode(&result); err != nil {
		return nil, err
	}

	resultJSON, err := bson.MarshalExtJSON(result, true, true)
	if err != nil {
		return nil, fmt.Errorf("explain: unable to encode explain result of %s: %s", a.params.Query, err)
	}

	return resultJSON, nil
}

func (a *mongodbExplainAction) sealed() {}
