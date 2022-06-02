package encryption

import (
	"context"
	_ "embed"
	"github.com/percona/pmm/utils/rsa_encryptor"
)

//go:embed default-key
var privateKey []byte

const EncryptorKey = "encryptor"

func NewFromDefaultKey() (*rsa_encryptor.Service, error) {
	return rsa_encryptor.NewFromPrivateKey("d1", privateKey)
}

func InjectEncryptorIfNotPresent(ctx context.Context) (context.Context, error) {
	encryptor := ctx.Value(EncryptorKey)
	if encryptor == nil {
		encryptor, err := NewFromDefaultKey()
		if err != nil {
			return nil, err
		}
		return context.WithValue(ctx, EncryptorKey, encryptor), nil
	}

	return ctx, nil
}

func GetEncryptor(ctx context.Context) *rsa_encryptor.Service {
	return ctx.Value(EncryptorKey).(*rsa_encryptor.Service)
}
