package common

import (
	"time"
)

type PGPString string
type TOTP string

type UserInfo struct {
	UserID        uint64
	AccountNumber string
	Timestamp     time.Time
	TOTP          TOTP
	PayLoad       PGPString
}

const AccountNumberLength = 10
