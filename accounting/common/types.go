package common

import (
	"regexp"
	"time"

	"git.condensat.tech/bank"
)

const (
	WithOperatorAuth = true
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

type BIC string

func (p *BIC) Valid() (bool, error) {
	valid, err := regexp.MatchString("([a-zA-Z]{4})([a-zA-Z]{2})(([2-9a-zA-Z]{1})([0-9a-np-zA-NP-Z]{1}))((([0-9a-wy-zA-WY-Z]{1})([0-9a-zA-Z]{2}))|([xX]{3})|)", string(*p))
	if err != nil {
		return false, err
	}

	return valid, nil
}

type AuthInfo struct {
	OperatorAccount string
	TOTP            TOTP
}

type CryptoWithdraw struct {
	WithdrawID uint64
	TargetID   uint64
	UserName   string
	Address    string
	Amount     float64
	Currency   string
}

type CryptoFetchPendingWithdrawList struct {
	PendingWithdraws []CryptoWithdraw
}
}

type FiatSepaInfo struct {
	IBAN
	BIC
	Label string
}

type FiatFetchPendingWithdraw struct {
	ID       uint64
	UserName string
	IBAN     string
	BIC      string
	Currency string
	Amount   float64
}

type FiatFetchPendingWithdrawList struct {
	PendingWithdraws []FiatFetchPendingWithdraw
}

type FiatFinalizeWithdraw struct {
	AuthInfo
	ID       uint64
	UserName string
	IBAN
	Currency string
	Amount   float64
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

func (p *AuthInfo) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AuthInfo) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *CryptoFetchPendingWithdrawList) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *CryptoFetchPendingWithdrawList) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *FiatWithdraw) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *FiatWithdraw) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *FiatFetchPendingWithdrawList) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *FiatFetchPendingWithdrawList) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *FiatFinalizeWithdraw) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *FiatFinalizeWithdraw) Decode(data []byte) error {
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
