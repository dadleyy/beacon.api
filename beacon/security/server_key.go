package security

import "fmt"
import "io/ioutil"
import "crypto/rsa"
import "crypto/x509"
import "encoding/pem"
import "encoding/hex"

// ServerKey objects contain the rsa private key used to secure communications w/ the api
type ServerKey struct {
	*rsa.PrivateKey
}

// SharedSecret returns the string version of the rsa public key
func (key *ServerKey) SharedSecret() (string, error) {
	publicKeyData, e := x509.MarshalPKIXPublicKey(key.Public())

	if e != nil {
		return "", e
	}

	return hex.EncodeToString(publicKeyData), nil
}

// ReadServerKeyFromFile returns a new device key from a filename
func ReadServerKeyFromFile(filename string) (*ServerKey, error) {
	privateKeyData, e := ioutil.ReadFile(filename)

	if e != nil {
		return nil, e
	}

	privateBlock, _ := pem.Decode(privateKeyData)

	if privateBlock == nil {
		return nil, fmt.Errorf("invalid-pem")
	}

	privateKey, e := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)

	if e != nil {
		return nil, e
	}

	return &ServerKey{PrivateKey: privateKey}, nil
}
