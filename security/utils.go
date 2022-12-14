package security

import (
	"context"

	"git.condensat.tech/bank/security/utils"
)

func GenerateSeed() ([SeedKeySize]byte, error) {
	var seed [SeedKeySize]byte
	err := utils.GenerateRand(seed[:])
	return seed, err
}

func GenerateNonce() ([NonceSize]byte, error) {
	var nonce [NonceSize]byte
	err := utils.GenerateRand(nonce[:])
	return nonce, err
}

func xorKey(ctx context.Context, secretKey SecretKey) SecretKey {
	salt := ctx.Value(KeyPrivateKeySalt).([]byte)
	if len(salt) < MinSaltSize {
		panic("Wrong salt size")
	}

	key := utils.HashBytes(salt)
	defer utils.Memzero(key)

	var result SecretKey
	utils.Xor(result[:], key, secretKey[:])
	return result
}

func xorHashSeed(ctx context.Context, hashSeedKey HashSeedKey) HashSeedKey {
	salt := ctx.Value(KeyPrivateKeySalt).([]byte)
	if len(salt) < MinSaltSize {
		panic("Wrong salt size")
	}

	key := utils.HashBytes(salt)
	defer utils.Memzero(key)

	var result HashSeedKey
	utils.Xor(result[:], key, hashSeedKey[:])
	return result
}
