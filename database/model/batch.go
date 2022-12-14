package model

import (
	"time"
)

type BatchID ID
type BatchData String
type BatchNetwork String

const (
	BatchNetworkSepa  BatchNetwork = "sepa"
	BatchNetworkSwift BatchNetwork = "swift"
	BatchNetworkCard  BatchNetwork = "card"

	BatchNetworkBitcoin          BatchNetwork = "bitcoin"
	BatchNetworkBitcoinTestnet   BatchNetwork = "bitcoin-testnet"
	BatchNetworkBitcoinLiquid    BatchNetwork = "liquid"
	BatchNetworkBitcoinLightning BatchNetwork = "lightning"
)

type Batch struct {
	ID           BatchID      `gorm:"primary_key"`
	Timestamp    time.Time    `gorm:"index;not null;type:timestamp"`   // Creation timestamp
	ExecuteAfter time.Time    `gorm:"index;not null;type:timestamp"`   // Execute after timestamp
	Capacity     Int          `gorm:"index;not null"`                  // Batch capacity
	Network      BatchNetwork `gorm:"index;not null;size:24"`          // Network [sepa, swift, card, bitcoin, bitcoin-testnet, liquid, lightning]
	Data         BatchData    `gorm:"type:blob;not null;default:'{}'"` // Batch data
}

func (p *Batch) IsComplete() bool {
	return !time.Now().UTC().Before(p.ExecuteAfter)
}
