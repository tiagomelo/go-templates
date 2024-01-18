package tlscreds

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
)

const (
	// caCert is the path to the CA's certificate file. This certificate
	// is used to verify the client's certificates.
	caCert = "cert/ca-cert.pem"

	// serverCert is the path to the server's certificate file.
	// This certificate is presented to clients during TLS handshake.
	serverCert = "cert/server-cert.pem"

	// serverKey is the path to the server's private key file.
	// This key is used for TLS encryption and must be kept secure.
	serverKey = "cert/server-key.pem"
)

// For ease of unit testing.
var (
	readFile           = os.ReadFile
	loadX509KeyPair    = tls.LoadX509KeyPair
	appendCertsFromPEM = func(certPool *x509.CertPool, pemCerts []byte) (ok bool) {
		return certPool.AppendCertsFromPEM(pemCerts)
	}
)

func New() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed client's certificate
	pemClientCA, err := readFile(caCert)
	if err != nil {
		return nil, errors.Wrap(err, "loading CA's certificate")
	}
	certPool := x509.NewCertPool()
	if !appendCertsFromPEM(certPool, pemClientCA) {
		return nil, errors.New("failed to add client CA's certificate")
	}
	// Load server's certificate and private key
	serverCert, err := loadX509KeyPair(serverCert, serverKey)
	if err != nil {
		return nil, errors.Wrap(err, "loading server's certificate and private key")
	}
	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}
	return credentials.NewTLS(config), nil
}
