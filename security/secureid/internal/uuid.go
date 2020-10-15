package internal

import (
	"github.com/google/uuid"
)

// ToUUID convert byte slice into uuid
// reverse order is performed
// data length must be 16 bytes
func ToUUID(data []byte) uuid.UUID {
	id, err := uuid.FromBytes(ReverseBytes(data[:]))
	if err != nil {
		return uuid.UUID{}
	}
	return id
}

// FromUUID return bytes from uuid string representation (see RFC 4122)
// reverse order is performed
func FromUUID(data string) []byte {
	id, err := uuid.Parse(data)
	if err != nil {
		return nil
	}

	return ReverseBytes(id[:])
}
