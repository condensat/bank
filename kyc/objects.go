package kyc

import (
	"git.condensat.tech/bank"
)

type KycStart struct {
	UserID uint64
	Email  string
}

func (p *KycStart) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *KycStart) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

type KycStartResponse struct {
	ID string
}

func (p *KycStartResponse) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *KycStartResponse) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
