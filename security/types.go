package security

import (
	"crypto/ed25519"
	"errors"

	sodium "git.condensat.tech/bank/security/internal/libsodium"
)

const (
	KeyPrivateKeySalt = "Key.PrivateKeySalt"
)

const (
	SignatureSecretKeySize = ed25519.PrivateKeySize
	SignaturePublicKeySize = ed25519.PublicKeySize

	EncryptionKeySize = sodium.Curve25519Size
	NonceSize         = 24
	MinSaltSize       = 24

	AuthenticationKeySize       = EncryptionKeySize
	AuthenticationKeyDigestSize = EncryptionKeySize

	SeedKeySize     = ed25519.SeedSize
	HashSeedKeySize = 32
)

var (
	ErrInvalidKey      = errors.New("Invalid key")
	ErrSignMessage     = errors.New("Message Sign Failed")
	ErrVerifySignature = errors.New("Signature verification failed")
	ErrNoSignature     = errors.New("No Signature found")
)

type SecretKey [SignatureSecretKeySize]byte
type SignatureSecretKey [SignatureSecretKeySize]byte
type SignaturePublicKey [SignaturePublicKeySize]byte

type EncryptionPublicKey [EncryptionKeySize]byte
type EncryptionPrivateKey [EncryptionKeySize]byte

type AuthenticationKey [AuthenticationKeySize]byte
type AuthenticationDigest [AuthenticationKeyDigestSize]byte

type SeedKey [SeedKeySize]byte
type HashSeedKey [HashSeedKeySize]byte
