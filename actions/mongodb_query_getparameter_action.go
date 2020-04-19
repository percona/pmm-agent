package actions

import (
	"context"

	"github.com/percona/pmm/api/agentpb"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongodbQueryGetparameterAction struct {
	id     string
	params *agentpb.StartActionRequest_MongoDBQueryGetParameterParams
}

// NewMongoDBQueryGetparameterAction creates a MongoDB getParameter query Action.
func NewMongoDBQueryGetparameterAction(id string, params *agentpb.StartActionRequest_MongoDBQueryGetParameterParams) Action {
	return &mongodbQueryGetparameterAction{
		id:     id,
		params: params,
	}
}

// ID returns an Action ID.
func (a *mongodbQueryGetparameterAction) ID() string {
	return a.id
}

// Type returns an Action type.
func (a *mongodbQueryGetparameterAction) Type() string {
	return "mongodb-query-getparameter"
}

// Run runs an Action and returns output and error.
func (a *mongodbQueryGetparameterAction) Run(ctx context.Context) ([]byte, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(a.params.Dsn))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer client.Disconnect(ctx) //nolint:errcheck

	runCommand := bson.D{{"getParameter", "*"}} //nolint:govet
	res := client.Database("admin").RunCommand(ctx, runCommand)

	var doc map[string]interface{}
	if err = res.Decode(&doc); err != nil {
		return nil, errors.WithStack(err)
	}

	data := []map[string]interface{}{doc}
	return agentpb.MarshalActionQueryResult(data)
}

func (a *mongodbQueryGetparameterAction) sealed() {}
