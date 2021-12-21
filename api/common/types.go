package common

import (
	"git.condensat.tech/bank"
)

type PGPPublicKey string

type AuthInfo struct {
	OperatorAccount string
	TOTP            TOTP
}

type UserCreation struct {
	AuthInfo
	PGPPublicKey PGPPublicKey
	UserInfo     UserInfo
}

func (p *UserCreation) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *UserCreation) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
