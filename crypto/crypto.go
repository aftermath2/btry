package crypto

import (
	"crypto/ed25519"
	"encoding/hex"

	"github.com/pkg/errors"
)

// ValidatePublicKey returns an error if the provided public key is invalid.
func ValidatePublicKey(publicKey string) error {
	if len(publicKey) != hex.EncodedLen(ed25519.PublicKeySize) {
		return errors.New("invalid public key length")
	}

	if _, err := hex.DecodeString(publicKey); err != nil {
		return errors.Wrap(err, "invalid public key encoding")
	}

	return nil
}

// VerifySignature validates the signature with the public key.
func VerifySignature(publicKey, signature string) error {
	pubKey, err := hex.DecodeString(publicKey)
	if err != nil {
		return errors.Wrap(err, "decoding public key")
	}

	sig, err := hex.DecodeString(signature)
	if err != nil {
		return errors.Wrap(err, "decoding signature")
	}

	if !ed25519.Verify(pubKey, pubKey, sig) {
		return errors.New("invalid signature")
	}

	return nil
}
