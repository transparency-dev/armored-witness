package kmssigner

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/pem"

	"cloud.google.com/go/kms/apiv1"

	"cloud.google.com/go/kms/apiv1/kmspb"
)

// KeyVersionNameFormat is the GCP resource identifier for a key version.
// google.cloud.kms.v1.CryptoKeyVersion.name
// https://cloud.google.com/php/docs/reference/cloud-kms/latest/V1.CryptoKeyVersion
const KeyVersionNameFormat = "projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s/cryptoKeyVersions/%d"

// Signer is an implementation of a
// [note signer](https://pkg.go.dev/golang.org/x/mod/sumdb/note#Signer) which
// interfaces with GCP KMS.
type Signer struct {
	// ctx must be stored because Signer is used as an implementation of the
	// note.Signer interface, which does not allow for a context in the Sign
	// method. However, the KMS AsymmetricSign API requires a context.
	ctx     context.Context
	client  *kms.KeyManagementClient
	keyHash uint32
	keyName string
}

// New creates a signer which uses keys in GCP KMS. The signing algorithm is
// expected to be
// [ECDSA](https://cloud.google.com/certificate-authority-service/docs/choosing-key-algorithm#ecdsa).
// To open a note signed by this Signer, the verifier must also be ECDSA.
func New(ctx context.Context, c *kms.KeyManagementClient, keyName string) (*Signer, error) {
	s := &Signer{}

	s.client = c
	s.ctx = ctx
	s.keyName = keyName

	// Set keyHash.
	req := &kmspb.GetPublicKeyRequest{
		Name: s.keyName,
	}
	resp, err := c.GetPublicKey(ctx, req)
	if err != nil {
		return nil, err
	}
	decoded, _ := pem.Decode([]byte(resp.Pem))

	// Calculate key hash from the checksum of the public key DER.
	checksum := sha256.Sum256(decoded.Bytes)
	s.keyHash = binary.BigEndian.Uint32(checksum[:])

	return s, nil
}

// Name identifies the key that this Signer uses.
func (s *Signer) Name() string {
	return s.keyName
}

// KeyHash returns the first 4 bytes of the SHA256 hash of the Signer's public
// key. It is used as a hint in identifying the correct key to verify with.
func (s *Signer) KeyHash() uint32 {
	return s.keyHash
}

// Sign returns a signature for the given message.
func (s *Signer) Sign(msg []byte) ([]byte, error) {
	req := &kmspb.AsymmetricSignRequest{
		Name: s.keyName,
		Data: msg,
	}
	resp, err := s.client.AsymmetricSign(s.ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetSignature(), nil
}
