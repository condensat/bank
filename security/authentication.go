package security

import (
	"git.condensat.tech/bank/security/utils"

	"golang.org/x/crypto/nacl/auth"
)

func AuthenticateMessage(authenticateKey AuthenticationKey, message []byte) AuthenticationDigest {
	defer utils.Memzero(authenticateKey[:])

	key := [AuthenticationKeySize]byte(authenticateKey)
	defer utils.Memzero(key[:])

	digest := auth.Sum(message, &key)
	defer utils.Memzero(digest[:])

	var auth AuthenticationDigest
	copy(auth[:], digest[:])
	return auth
}

func VerifyMessageAuthentication(authenticateKey AuthenticationKey, digest AuthenticationDigest, message []byte) bool {
	defer utils.Memzero(authenticateKey[:])

	key := [AuthenticationKeySize]byte(authenticateKey)
	defer utils.Memzero(key[:])

	return auth.Verify(digest[:], message, &key)
}
