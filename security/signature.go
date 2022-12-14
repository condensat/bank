package security

import (
	"git.condensat.tech/bank"

	"git.condensat.tech/bank/security/utils"
	"golang.org/x/crypto/nacl/sign"
)

func Sign(secretKey SignatureSecretKey, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, bank.ErrNoData
	}
	defer utils.Memzero(secretKey[:])

	var privateKey [SignatureSecretKeySize]byte
	defer utils.Memzero(privateKey[:])
	copy(privateKey[:], secretKey[:])

	s := sign.Sign(nil, data, &privateKey)
	if len(s) == 0 {
		return nil, ErrSignMessage
	}

	return s, nil
}

func VerifySignature(publicKey SignaturePublicKey, signedMessage []byte) (bool, error) {
	if len(signedMessage) <= sign.Overhead {
		return false, ErrNoSignature
	}
	defer utils.Memzero(publicKey[:])

	var pubKey [SignaturePublicKeySize]byte
	defer utils.Memzero(pubKey[:])
	copy(pubKey[:], publicKey[:])

	_, ok := sign.Open(nil, signedMessage, &pubKey)
	if !ok {
		return false, ErrVerifySignature
	}

	return true, nil
}
