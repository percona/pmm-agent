package mongo_fix

import (
	"net/url"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ClientForDSN applies URI to Client
func ClientForDSN(dsn string) (*options.ClientOptions, error) {
	parsedDsn, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse DSN")
	}

	clientOptions := options.Client().ApplyURI(dsn)

	// Workaround for PMM-9320
	username := parsedDsn.User.Username()
	password, _ := parsedDsn.User.Password()
	if username != "" || password != "" {
		clientOptions = clientOptions.SetAuth(options.Credential{Username: username, Password: password})
	}

	return clientOptions, nil
}
