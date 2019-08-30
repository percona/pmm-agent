package actions

import (
	"context"
	"fmt"
	"testing"

	"github.com/percona/percona-toolkit/src/go/mongolib/proto"
	"github.com/percona/pmm-agent/utils/tests"
	"github.com/percona/pmm/api/agentpb"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		Dsn:      tests.MongoDBDSN(),
		Database: database,
		Query:    string(buf),
	}

	ex := NewNomgoDBExplain(id, params)
	_, err = ex.Run(ctx)
	assert.Nil(t, err)
	client.Database(database).Drop(ctx)
}

func prepareData(ctx context.Context, client *mongo.Client, database, collection string) error {
	max := int64(100)
	count, _ := client.Database(database).Collection(collection).CountDocuments(ctx, nil)

	if count < max {
		for i := int64(0); i < max; i++ {
			client.Database(database).Collection(collection).InsertOne(ctx, primitive.M{"f1": i, "f2": fmt.Sprintf("text_%5d", max-i)})
		}
	}

	return nil
}
