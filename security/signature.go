package security

import (
	"crypto/ed25519"
	"crypto/sha256"

	"git.condensat.tech/bank"
)

func Sign(key bank.SharedKey, data []byte) ([]byte, error) {
	if !IsKeyValid(key) {
		return nil, ErrInvalidKey
	}
	if len(data) == 0 {
		return nil, bank.ErrNoData
	}

	hash := sha256.Sum256(key[:])
	priv := ed25519.NewKeyFromSeed(hash[:])

	return ed25519.Sign(priv, data), nil
}

func Verify(key bank.SharedKey, data, signature []byte) bool {
	if !IsKeyValid(key) {
		return false
	}
	if len(data) == 0 {
		return false
	}
	if len(signature) != ed25519.SignatureSize {
		return false
	}

	hash := sha256.Sum256(key[:])
	priv := ed25519.NewKeyFromSeed(hash[:])

	pub := priv.Public().(ed25519.PublicKey)
	return ed25519.Verify(pub, data, signature)
}
