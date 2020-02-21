package monitor

import (
	"git.condensat.tech/bank"
)

func (p *ProcessInfo) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *ProcessInfo) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
