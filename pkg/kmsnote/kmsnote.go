package kmsnote

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"

	kms "cloud.google.com/go/kms/apiv1"
	"golang.org/x/mod/sumdb/note"

	"cloud.google.com/go/kms/apiv1/kmspb"
)

const (
	// KeyVersionNameFormat is the GCP resource identifier for a key version.
	// google.cloud.kms.v1.CryptoKeyVersion.name
	// https://cloud.google.com/php/docs/reference/cloud-kms/latest/V1.CryptoKeyVersion
	KeyVersionNameFormat = "projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s/cryptoKeyVersions/%d"
	// From
	// https://cs.opensource.google/go/x/mod/+/refs/tags/v0.12.0:sumdb/note/note.go;l=232;drc=baa5c2d058db25484c20d76985ba394e73176132
	algEd25519 = 1
)

func publicKeyFromPEM(pemKey []byte) ([]byte, error) {
	block, _ := pem.Decode(pemKey)
	if block == nil {
		return nil, errors.New("failed to decode pemKey")
	}

	k, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := k.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("failed to assert ed25519.PublicKey type")
	}

	return publicKey, nil
}

// keyHash calculates the ed25519 key hash from the key name and public key.
func keyHash(keyName string, publicKey []byte) (uint32, error) {
	h := sha256.New()
	h.Write([]byte(keyName))
	h.Write([]byte("\n"))
	prefixedPublicKey := append([]byte{algEd25519}, publicKey...)
	h.Write(prefixedPublicKey)
	sum := h.Sum(nil)

	return binary.BigEndian.Uint32(sum), nil
}

// Signer is an implementation of a
// [note signer](https://pkg.go.dev/golang.org/x/mod/sumdb/note#Signer) which
// interfaces with GCP KMS.
type Signer struct {
	// ctx must be stored because Signer is used as an implementation of the
	// note.Signer interface, which does not allow for a context in the Sign
	// method. However, the KMS AsymmetricSign API requires a context.
	ctx         context.Context
	client      *kms.KeyManagementClient
	keyHash     uint32
	keyName     string
	keyResource string
}

// NewSigner creates a signer which uses keys in GCP KMS. The signing algorithm
// is expected to be
// [Ed25519](https://pkg.go.dev/golang.org/x/mod/sumdb/note#hdr-Generating_Keys).
// To open a note signed by this Signer, the verifier must also be Ed25519.
func NewSigner(ctx context.Context, c *kms.KeyManagementClient, keyName string) (*Signer, error) {
	s := &Signer{}

	s.client = c
	s.ctx = ctx
	s.keyName = noteSignerName
	s.keyResource = keyResource

	// Set keyHash.
	req := &kmspb.GetPublicKeyRequest{
		Name: keyResource,
	}
	resp, err := c.GetPublicKey(ctx, req)
	if err != nil {
		return nil, err
	}

	publicKey, err := publicKeyFromPEM([]byte(resp.Pem))
	if err != nil {
		return nil, err
	}

	kh, err := keyHash(s.keyName, publicKey)
	if err != nil {
		return nil, err
	}
	s.keyHash = kh

	return s, nil
}

// Name identifies the key that this Signer uses.
func (s *Signer) Name() string {
	return s.keyName
}

// KeyHash returns the computed key hash of the signer's public key and name.
// It is used as a hint in identifying the correct key to verify with.
func (s *Signer) KeyHash() uint32 {
	return s.keyHash
}

// Sign returns a signature for the given message.
func (s *Signer) Sign(msg []byte) ([]byte, error) {
	req := &kmspb.AsymmetricSignRequest{
		Name: s.keyResource,
		Data: msg,
	}
	resp, err := s.client.AsymmetricSign(s.ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetSignature(), nil
}

// NewVerifier creates a verifier which uses keys in GCP KMS. The signing
// algorithm is expected to be
// [Ed25519](https://pkg.go.dev/golang.org/x/mod/sumdb/note#hdr-Generating_Keys).
func NewVerifier(ctx context.Context, c *kms.KeyManagementClient, kmsKeyName, noteKeyName string) (note.Verifier, error) {
	req := &kmspb.GetPublicKeyRequest{
		Name: kmsKeyName,
	}
	resp, err := c.GetPublicKey(ctx, req)
	if err != nil {
		return nil, err
	}

	publicKey, err := publicKeyFromPEM([]byte(resp.Pem))
	if err != nil {
		return nil, err
	}

	h, err := keyHash(noteKeyName, publicKey)
	if err != nil {
		return nil, err
	}

	prefixedPublicKey := append([]byte{algEd25519}, publicKey...)
	vkey := fmt.Sprintf("%s+%08x+%s", noteKeyName, h, base64.StdEncoding.EncodeToString(prefixedPublicKey))
	return note.NewVerifier(vkey)
}
