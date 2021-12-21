package security

import (
	"testing"

	"git.condensat.tech/bank/security/utils"
	"github.com/shengdoushi/base58"
	"golang.org/x/net/context"
)

func prepareContext() context.Context {
	ctx := context.Background()

	var seed [HashSeedKeySize]byte
	err := utils.GenerateRand(seed[:])
	if err != nil {
		panic(err)
	}

	ctx = context.WithValue(ctx, KeyPrivateKeySalt, utils.GenerateRandN(32))
	ctx = PasswordHashSeedContext(ctx, base58.Encode(seed[:], DefaultAlphabet))

	return ctx
}

func TestReadWriteSecret(t *testing.T) {
	ctx := prepareContext()

	message := "Hello, Secret!"

	got := ReadSecret(ctx, WriteSecret(ctx, message))
	if got != message {
		t.Errorf(got)
	}
}
