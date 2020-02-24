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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/percona/percona-toolkit/src/go/mongolib/proto"
	"github.com/percona/pmm/api/agentpb"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/percona/pmm-agent/utils/tests"
)

func TestMongoDBExplain(t *testing.T) {
	database := "test"
	collection := "test_col"
	id := "abcd1234"
	ctx := context.TODO()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(tests.MongoDBDSN()))
	if err != nil {
		t.Fatalf("Cannot connect to MongoDB: %s", err)
	}

	if err := prepareData(ctx, client, database, collection); err != nil {
		t.Fatalf("Cannot prepare MongoDB testing data: %s", err)
	}

	eq := proto.ExampleQuery{
		Ns: "test.coll",
		Op: "query",
		Query: proto.BsonD{
			{
				Key: "k",
				Value: proto.BsonD{
					{
						Key:   "$lte",
						Value: int32(1),
					},
				},
			},
		},
		Command:            nil,
		OriginatingCommand: nil,
		UpdateObj:          nil,
	}
	buf, _ := bson.MarshalExtJSON(eq, true, true)

	params := &agentpb.StartActionRequest_MongoDBExplainParams{
		Dsn:   tests.MongoDBDSN(),
		Query: string(buf),
	}

	ex := NewMongoDBExplain(id, params)
	res, err := ex.Run(ctx)
	assert.Nil(t, err)

	// explain package has a lot of tests for different queries and different MongoDB versions.
	// Here we only need to check we are receiving a valid response
	explainM := make(map[string]interface{})
	err = json.Unmarshal(res, &explainM)
	assert.Nil(t, err)
	queryPlanner, ok := explainM["queryPlanner"]
	assert.Equal(t, ok, true)
	assert.NotEmpty(t, queryPlanner)

	if err := client.Database(database).Drop(ctx); err != nil {
		t.Errorf("Cannot drop testing database for cleanup")
	}
}

func prepareData(ctx context.Context, client *mongo.Client, database, collection string) error {
	max := int64(100)
	count, _ := client.Database(database).Collection(collection).CountDocuments(ctx, nil)

	if count < max {
		for i := int64(0); i < max; i++ {
			doc := primitive.M{"f1": i, "f2": fmt.Sprintf("text_%5d", max-i)}
			if _, err := client.Database(database).Collection(collection).InsertOne(ctx, doc); err != nil {
				return err
			}
		}
	}

	return nil
}
