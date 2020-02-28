package common

import (
	"time"

	"git.condensat.tech/bank"
)

type StackListService struct {
	Since    time.Duration
	Services []string
}

func (p *StackListService) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *StackListService) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
