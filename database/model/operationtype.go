package model

import "errors"

type OperationType String

const (
	OperationTypeInvalid OperationType = ""

	OperationTypeInit       OperationType = "init"
	OperationTypeDeposit    OperationType = "deposit"
	OperationTypeWithdraw   OperationType = "withdraw"
	OperationTypeTransfer   OperationType = "transfer"
	OperationTypeAdjustment OperationType = "adjustment"

	OperationTypeNone  OperationType = "none"
	OperationTypeOther OperationType = "other"
)

var (
	ErrOperationTypeInvalid = errors.New("Invalid OperationType")
)

func (p OperationType) Valid() bool {
	switch p {
	case OperationTypeInit:
		fallthrough
	case OperationTypeDeposit:
		fallthrough
	case OperationTypeWithdraw:
		fallthrough
	case OperationTypeTransfer:
		fallthrough
	case OperationTypeAdjustment:
		fallthrough

	case OperationTypeNone:
		fallthrough
	case OperationTypeOther:
		return true

	default:
		return false
	}
}

func ParseOperationType(str string) OperationType {
	ret := OperationType(str)
	if !ret.Valid() {
		return OperationTypeInvalid
	}
	return ret
}

func (p OperationType) String() string {
	if !p.Valid() {
		return string(OperationTypeInvalid)
	}
	return string(p)
}

func knownOperationType() []OperationType {
	return []OperationType{
		OperationTypeInvalid,

		OperationTypeInit,
		OperationTypeDeposit,
		OperationTypeWithdraw,
		OperationTypeTransfer,
		OperationTypeAdjustment,

		OperationTypeNone,
		OperationTypeOther,
	}
}
