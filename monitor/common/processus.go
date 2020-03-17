package common

import (
	"time"

	"git.condensat.tech/bank"
)

type ProcessInfo struct {
	ID        uint64    `gorm:"primary_key"`
	Timestamp time.Time `gorm:"index;not null"`
	AppName   string    `gorm:"index;not null"`
	Hostname  string    `gorm:"index;not null"`
	PID       int       `gorm:"not null"`

	MemAlloc      uint64 `gorm:"not null"`
	MemTotalAlloc uint64 `gorm:"not null"`
	MemSys        uint64 `gorm:"not null"`
	MemLookups    uint64 `gorm:"not null"`

	NumCPU       uint64  `gorm:"not null"`
	NumGoroutine uint64  `gorm:"not null"`
	NumCgoCall   uint64  `gorm:"not null"`
	CPUUsage     float64 `gorm:"not null"`
}

func (p *ProcessInfo) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *ProcessInfo) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
