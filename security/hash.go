package security

import (
	"context"

	"crypto/subtle"

	"git.condensat.tech/bank/security/utils"
)

const (
	KeyHasherWorker = "Security.KeyHasherWorker"
)

// SaltedHash return argon2 key from salt and input
// do not store or reuse salt
func SaltedHash(ctx context.Context, password []byte) []byte {
	if len(password) == 0 {
		panic("Invalid Input")
	}
	salt := PasswordHashSalt(ctx)
	defer utils.Memzero(salt[:])

	worker := ctx.Value(KeyHasherWorker).(*HasherWorker)
	if worker == nil {
		panic("Invalid HasherWorker")
	}

	return worker.doHash(salt[:], password)
}

// SaltedHashVerify check if hash correspond to argon2 hash from salt and input
// do not store or reuse salt
func SaltedHashVerify(ctx context.Context, password []byte, hash []byte) bool {
	if len(password) == 0 {
		return false
	}
	salt := PasswordHashSalt(ctx)
	defer utils.Memzero(salt[:])

	worker := ctx.Value(KeyHasherWorker).(*HasherWorker)
	if worker == nil {
		panic("Invalid HasherWorker")
	}

	// compute argon hash
	dk := worker.doHash(salt[:], password)
	defer utils.Memzero(dk[:])

	// check if length match
	if subtle.ConstantTimeEq(int32(len(dk)), int32(len(hash))) == 0 {
		return false
	}

	// Compare keys content
	return subtle.ConstantTimeCompare(dk, hash) == 1
}
