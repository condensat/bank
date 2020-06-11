package common

import (
	"git.condensat.tech/bank"
)

type CryptoAddress struct {
	CryptoAddressID uint64
	Chain           string
	AccountID       uint64
	PublicAddress   string
	Unconfidential  string
}

type TransactionInfo struct {
	Chain         string
	Account       string
	Address       string
	Asset         string
	TxID          string
	Vout          int64
	Amount        float64
	Confirmations int64
	Spendable     bool
}

type AddressInfo struct {
	Chain          string
	PublicAddress  string
	Unconfidential string
}

func (p *CryptoAddress) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *CryptoAddress) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}