package tls_helpers

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/go-sql-driver/mysql"
)

func RegisterMySQL(files map[string]string) error {
	ca := x509.NewCertPool()
	cert, err := tls.X509KeyPair([]byte(files["tlsCert"]), []byte(files["tlsKey"]))
	if err != nil {
		return err
	}

	if ok := ca.AppendCertsFromPEM([]byte(files["tlsCa"])); ok {
		err = mysql.RegisterTLSConfig("custom", &tls.Config{
			InsecureSkipVerify: false,
			RootCAs:            ca,
			Certificates:       []tls.Certificate{cert},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
