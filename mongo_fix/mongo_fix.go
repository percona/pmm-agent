package mongo_fix

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/url"
)

// ClientForDSN applies URI to Client
func ClientForDSN(dsn string) (*options.ClientOptions, error) {
	parsedDsn, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse DSN")
	}

	username := parsedDsn.User.Username()
	password, _ := parsedDsn.User.Password()
	creds := options.Credential{Username: username, Password: password}

	clientOptions := options.Client().
		ApplyURI(dsn).
		SetAuth(creds) //Workaround for PMM-9320

	return clientOptions, nil
}
