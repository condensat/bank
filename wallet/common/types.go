package common

import (
	"git.condensat.tech/bank"
)

type CryptoMode string

const (
	CryptoModeBitcoinCore CryptoMode = "bitcoin-core"
	CryptoModeCryptoSsm   CryptoMode = "crypto-ssm"
)

type CryptoAddress struct {
	CryptoAddressID  uint64
	Chain            string
	AccountID        uint64
	PublicAddress    string
	Unconfidential   string
	IgnoreAccounting bool
}

type SsmAddress struct {
	Chain       string
	Address     string
	PubKey      string
	BlindingKey string
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
	IsValid        bool
}

type UTXOInfo struct {
	TxID string
	Vout int
}

type SpendInfo struct {
	PublicAddress string
	Amount        float64
}

type SpendTx struct {
	TxID string
}

func (p *CryptoAddress) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *CryptoAddress) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AddressInfo) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AddressInfo) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
