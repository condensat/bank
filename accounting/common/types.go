package common

import (
	"regexp"
	"time"

	"git.condensat.tech/bank"
)

func Timestamp() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

type TOTP string

type IBAN string

func (p *IBAN) Valid() (bool, error) {
	valid, err := regexp.MatchString("[a-zA-Z]{2}[0-9]{2}[a-zA-Z0-9]{4}[0-9]{7}([a-zA-Z0-9]?){0,16}", string(*p))
	if err != nil {
		return false, err
	}

	return valid, nil
}

type AuthInfo struct {
	OperatorAccount string
	TOTP            TOTP
}

type FiatSepaInfo struct {
	IBAN
	BIC   string
	Label string
}

type FiatWithdraw struct {
	UserId      uint64
	Source      AccountEntry
	Destination FiatSepaInfo
}

type FiatDeposit struct {
	AuthInfo
	UserName    string
	Destination AccountEntry
}

type CurrencyType int

type CurrencyInfo struct {
	Name             string
	DisplayName      string
	DatabaseName     string
	Available        bool
	AutoCreate       bool
	Crypto           bool
	Type             CurrencyType
	Asset            bool
	DisplayPrecision uint
}

type CurrencyList struct {
	Currencies []CurrencyInfo
}

type AccountInfo struct {
	Timestamp   time.Time
	AccountID   uint64
	UserID      uint64
	Currency    CurrencyInfo
	Name        string
	Status      string
	Balance     float64
	TotalLocked float64
}

type AccountCreation struct {
	UserID uint64
	Info   AccountInfo
}

type UserAccounts struct {
	UserID uint64

	Accounts []AccountInfo
}

type AccountEntry struct {
	OperationID     uint64
	OperationPrevID uint64

	AccountID        uint64
	Currency         string
	ReferenceID      uint64
	OperationType    string
	SynchroneousType string

	Timestamp time.Time
	Label     string
	Amount    float64
	Balance   float64

	LockAmount  float64
	TotalLocked float64
}

type AccountTransfer struct {
	Source      AccountEntry
	Destination AccountEntry
}

type AccountHistory struct {
	AccountID   uint64
	DisplayName string
	Ticker      string
	From        time.Time
	To          time.Time

	Entries []AccountEntry
}

type CryptoTransfert struct {
	Chain     string
	PublicKey string
}

type AccountTransferWithdraw struct {
	BatchMode string
	Source    AccountEntry
	Crypto    CryptoTransfert
}

type WithdrawInfo struct {
	WithdrawID uint64
	Timestamp  time.Time
	AccountID  uint64
	Amount     float64
	Chain      string
	PublicKey  string
	Status     string
}

type UserWithdraws struct {
	UserID    uint64
	Withdraws []WithdrawInfo
}

type BatchWithdraw struct {
	BatchID       uint64
	BankAccountID uint64
	Network       string
	Status        string
	TxID          string
	Withdraws     []WithdrawInfo
}

type BatchWithdraws struct {
	Network string
	Batches []BatchWithdraw
}

type BatchStatus struct {
	BatchID uint64
	Status  string
}

type BatchUpdate struct {
	BatchStatus
	TxID   string
	Height int
}

func (p *FiatWithdraw) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *FiatWithdraw) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *FiatDeposit) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *FiatDeposit) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *CurrencyList) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *CurrencyList) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *CurrencyInfo) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *CurrencyInfo) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AccountInfo) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AccountInfo) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AccountCreation) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AccountCreation) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *UserAccounts) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *UserAccounts) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AccountEntry) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AccountEntry) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AccountTransfer) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AccountTransfer) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AccountHistory) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AccountHistory) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AccountTransferWithdraw) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AccountTransferWithdraw) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *WithdrawInfo) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *WithdrawInfo) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *UserWithdraws) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *UserWithdraws) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *BatchWithdraw) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *BatchWithdraw) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *BatchWithdraws) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *BatchWithdraws) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *BatchStatus) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *BatchStatus) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *BatchUpdate) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *BatchUpdate) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
