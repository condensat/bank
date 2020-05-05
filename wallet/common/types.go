package common

import (
	"git.condensat.tech/bank"
)

type CryptoAddress struct {
	Chain         string
	AccountID     uint64
	PublicAddress string
}

type TransactionInfo struct {
	Chain         string
	Account       string
	Address       string
	TxID          string
	Amount        float64
	Confirmations int64
	Spendable     bool
}

func (p *CryptoAddress) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *CryptoAddress) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
