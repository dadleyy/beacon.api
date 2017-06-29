package security

import "os"
import "io"
import "fmt"
import "log"
import "crypto/rsa"
import "crypto/x509"
import "crypto/rand"
import "crypto/sha256"
import "encoding/hex"

import "github.com/dadleyy/beacon.api/beacon/defs"

// DeviceKey implements the Signer interface that is used to encode messages sent to the device
type DeviceKey struct {
	*rsa.PublicKey
	*log.Logger
}

// Sign implements the signer interface
func (key *DeviceKey) Sign(out io.Writer, data []byte) error {
	signedData, e := rsa.EncryptOAEP(sha256.New(), rand.Reader, key.PublicKey, data, []byte("beacon"))

	if e != nil {
		return nil
	}

	_, e = out.Write(signedData)
	return e
}

// ParseDeviceKey returns a parsed device key capable of encoding device messages from a hex encoded byte array
func ParseDeviceKey(data string) (*DeviceKey, error) {
	block, e := hex.DecodeString(data)

	if e != nil {
		return nil, e
	}

	publicKey, e := x509.ParsePKIXPublicKey(block)

	if e != nil {
		return nil, e
	}

	rsaPublic, ok := publicKey.(*rsa.PublicKey)

	if ok != true {
		return nil, fmt.Errorf("invalid-public")
	}

	logger := log.New(os.Stdout, defs.DeviceControlLogPrefix, defs.DefaultLoggerFlags)
	return &DeviceKey{rsaPublic, logger}, nil
}
