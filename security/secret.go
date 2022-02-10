package security

import (
	"git.condensat.tech/bank/security/utils"
	"github.com/shengdoushi/base58"

	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/net/context"
)

func WriteSecret(ctx context.Context, message string) string {
	salt := PasswordHashSalt(ctx)
	defer utils.Memzero(salt[:])
	var secretKey [32]byte
	copy(secretKey[:], salt[:])
	nonce, _ := GenerateNonce()

	data := secretbox.Seal(nonce[:], []byte(message), &nonce, &secretKey)

	return base58.Encode(data, DefaultAlphabet)
}

func ReadSecret(ctx context.Context, message string) string {
	salt := PasswordHashSalt(ctx)
	defer utils.Memzero(salt[:])
	var secretKey [32]byte
	copy(secretKey[:], salt[:])

	encrypted, err := base58.Decode(message, DefaultAlphabet)
	if err != nil {
		encrypted = []byte(message)
		panic(err)
	}
	var decryptNonce [24]byte
	copy(decryptNonce[:], encrypted[:24])

	decrypted, ok := secretbox.Open(nil, encrypted[24:], &decryptNonce, &secretKey)
	if !ok {
		return ""
	}

	return string(decrypted)
}
