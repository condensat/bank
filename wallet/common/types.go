package common

import (
	"git.condensat.tech/bank"
)

type CryptoMode string
type AssetIssuanceMode string

const (
	CryptoModeBitcoinCore CryptoMode = "bitcoin-core"
	CryptoModeCryptoSsm   CryptoMode = "crypto-ssm"
)

const (
	AssetIssuanceModeWithAsset             AssetIssuanceMode = "asset-only"
	AssetIssuanceModeWithToken             AssetIssuanceMode = "with-token"
	AssetIssuanceModeWithContract          AssetIssuanceMode = "with-contract"
	AssetIssuanceModeWithTokenWithContract AssetIssuanceMode = "with-token-and-contract"
)

const (
	ElementsRegtestHash string = "b2e15d0d7a0c94e4e2ce0fe6e8691b9e451377f6e46e8045a86f7c4b5d4f0f23"
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
	TxID   string
	Vout   int
	Asset  string
	Amount float64
	Locked bool
}

type SpendAssetInfo struct {
	Hash          string
	ChangeAddress string
	ChangeAmount  float64
}

type SpendInfo struct {
	PublicAddress string
	Amount        float64
	// Asset optional
	Asset SpendAssetInfo
}

type SpendTx struct {
	TxID string
}

type IssuanceRequest struct {
	Chain              string            // mainly elements-regtest or LiquidV1 now, but can be useful for other chains later
	IssuerID           uint64            // User ID used for communication with our db
	Mode               AssetIssuanceMode // Issue an asset either with a reissuance token, a contract hash or both
	BlindIssuance      bool              // Issuance can be blinded or not
	AssetPublicAddress string            // Address we send the newly issued asset to
	AssetIssuedAmount  float64           // Max 21_000_000.0, but can be reissued many times

	// Optional
	TokenPublicAddress string  // Address we send the reissuance token to
	TokenIssuedAmount  float64 // I'd recommend it to be either 0 or 0.00000001 (1 sat)
	ContractHash       string  // 32B hash we can commit directly inside the asset ID
}

type IssuanceResponse struct {
	Chain     string   // mainly elements-regtest or LiquidV1 now, but can be useful for other chains later
	IssuerID  uint64   // User ID used for communication with our db
	AssetID   string   // This is the hex 64B Identifier of the asset. It is computed determinastically from a txid, a vout and an optional contract hash
	TokenID   string   // hex 64B identifier of the token that allows to reissue the asset
	TxID      string   // ID of the issuance transaction
	Vin       UTXOInfo // Txid and vout of the input the issuance is hooked to, used to compute asset ID with contract hash if any
	AssetVout int      // The vout of the new asset
	TokenVout int      // The vout of the token. We need this for reissuance
	Entropy   string   // Entropy is calculated with the issuance vin and contract hash if any
}
type WalletInfo struct {
	Chain  string
	Height int
	UTXOs  []UTXOInfo
}

type WalletStatus struct {
	Wallets []WalletInfo
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

func (p *WalletInfo) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *WalletInfo) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *WalletStatus) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *WalletStatus) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *IssuanceRequest) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *IssuanceRequest) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *IssuanceResponse) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *IssuanceResponse) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *IssuanceRequest) IsValid() bool {
	switch p.Mode {
	case AssetIssuanceModeWithAsset:
		if len(p.AssetPublicAddress) == 0 {
			return false
		}
		if p.AssetIssuedAmount <= 0.0 {
			return false
		}
		if len(p.TokenPublicAddress) != 0 {
			return false
		}
		if p.TokenIssuedAmount > 0.0 {
			return false
		}
		if len(p.ContractHash) != 0 {
			return false
		}
		return true
	case AssetIssuanceModeWithToken:
		if len(p.AssetPublicAddress) == 0 {
			return false
		}
		if p.AssetIssuedAmount <= 0.0 {
			return false
		}
		if len(p.TokenPublicAddress) == 0 {
			return false
		}
		if p.TokenIssuedAmount <= 0.0 {
			return false
		}
		if len(p.ContractHash) != 0 {
			return false
		}
		return true
	case AssetIssuanceModeWithContract:
		if len(p.AssetPublicAddress) == 0 {
			return false
		}
		if p.AssetIssuedAmount <= 0.0 {
			return false
		}
		if len(p.TokenPublicAddress) != 0 {
			return false
		}
		if p.TokenIssuedAmount > 0.0 {
			return false
		}
		if len(p.ContractHash) == 0 {
			return false
		}
		return true
	case AssetIssuanceModeWithTokenWithContract:
		if len(p.AssetPublicAddress) == 0 {
			return false
		}
		if p.AssetIssuedAmount <= 0.0 {
			return false
		}
		if len(p.TokenPublicAddress) == 0 {
			return false
		}
		if p.TokenIssuedAmount <= 0.0 {
			return false
		}
		if len(p.ContractHash) == 0 {
			return false
		}
		return true
	default:
		return false
	}
}
