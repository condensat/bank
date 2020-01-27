package security

import (
	"context"

	"crypto/subtle"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
)

// SaltedHash return argon2 key from salt and input
func SaltedHash(ctx context.Context, password, salt []byte) []byte {
	if len(password) == 0 || len(salt) < 8 {
		logger.Logger(ctx).
			Panic("Invalid Input")
	}

	worker := appcontext.HasherWorker(ctx).(*HasherWorker)
	if worker == nil {
		logger.Logger(ctx).
			Panic("Invalid HasherWorker")
	}

	return worker.doHash(password, salt)
}

// SaltedHashVerify check if key correspond to argon2 hash from salt and input
func SaltedHashVerify(ctx context.Context, password, salt []byte, key []byte) bool {
	if len(password) == 0 || len(salt) < 8 {
		return false
	}

	worker := appcontext.HasherWorker(ctx).(*HasherWorker)
	if worker == nil {
		logger.Logger(ctx).
			Panic("Invalid HasherWorker")
	}

	// compute argon hash
	dk := worker.doHash(password, salt)

	// check if length match
	if subtle.ConstantTimeEq(int32(len(dk)), int32(len(key))) == 0 {
		return false
	}

	// Compare keys content
	return subtle.ConstantTimeCompare(dk, key) == 1
}
