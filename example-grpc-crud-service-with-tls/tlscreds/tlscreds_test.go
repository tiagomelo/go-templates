package tlscreds

import (
	"crypto/tls"
	"crypto/x509"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name                   string
		mockReadFile           func(name string) ([]byte, error)
		mockLoadX509KeyPair    func(certFile string, keyFile string) (tls.Certificate, error)
		mockAppendCertsFromPEM func(certPool *x509.CertPool, pemCerts []byte) bool
		expectedError          error
	}{
		{
			name: "happy path",
			mockReadFile: func(name string) ([]byte, error) {
				return []byte{}, nil
			},
			mockLoadX509KeyPair: func(certFile, keyFile string) (tls.Certificate, error) {
				return tls.Certificate{}, nil
			},
			mockAppendCertsFromPEM: func(certPool *x509.CertPool, pemCerts []byte) bool {
				return true
			},
		},
		{
			name: "error when reading file",
			mockReadFile: func(name string) ([]byte, error) {
				return nil, errors.New("read file error")
			},
			expectedError: errors.New("loading CA's certificate: read file error"),
		},
		{
			name: "error when appending certs from pem",
			mockReadFile: func(name string) ([]byte, error) {
				return []byte{}, nil
			},
			mockAppendCertsFromPEM: func(certPool *x509.CertPool, pemCerts []byte) bool {
				return false
			},
			expectedError: errors.New("failed to add client CA's certificate"),
		},
		{
			name: "error when loading x509 key pair",
			mockReadFile: func(name string) ([]byte, error) {
				return []byte{}, nil
			},
			mockAppendCertsFromPEM: func(certPool *x509.CertPool, pemCerts []byte) bool {
				return true
			},
			mockLoadX509KeyPair: func(certFile, keyFile string) (tls.Certificate, error) {
				return tls.Certificate{}, errors.New("load error")
			},
			expectedError: errors.New("loading server's certificate and private key: load error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			readFile = tc.mockReadFile
			loadX509KeyPair = tc.mockLoadX509KeyPair
			appendCertsFromPEM = tc.mockAppendCertsFromPEM
			tls, err := New()
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.NotNil(t, tls)
			}
		})
	}
}
