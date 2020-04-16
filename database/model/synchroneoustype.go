package model

import (
	"errors"
)

type SynchroneousType String

const (
	SynchroneousTypeInvalid SynchroneousType = ""

	SynchroneousTypeSync       SynchroneousType = "sync"
	SynchroneousTypeAsyncStart SynchroneousType = "async-start"
	SynchroneousTypeAsyncEnd   SynchroneousType = "async-end"
)

var (
	ErrSynchroneousTypeInvalid = errors.New("Invalid SynchroneousType")
)

func (p SynchroneousType) Valid() bool {
	switch p {
	case SynchroneousTypeSync:
		fallthrough
	case SynchroneousTypeAsyncStart:
		fallthrough
	case SynchroneousTypeAsyncEnd:
		return true

	default:
		return false
	}
}

func ParseSynchroneousType(str string) SynchroneousType {
	ret := SynchroneousType(str)
	if !ret.Valid() {
		return SynchroneousTypeInvalid
	}
	return ret
}

func (p SynchroneousType) String() string {
	if !p.Valid() {
		return string(SynchroneousTypeInvalid)
	}
	return string(p)
}

func knownSynchroneousType() []SynchroneousType {
	return []SynchroneousType{
		SynchroneousTypeInvalid,

		SynchroneousTypeSync,
		SynchroneousTypeAsyncStart,
		SynchroneousTypeAsyncEnd,
	}
}
