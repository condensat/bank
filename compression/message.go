package compression

import (
	"errors"

	"git.condensat.tech/bank"
)

var (
	ErrOperationNotPermited = errors.New("Operation Not Permited")
)

func CompressMessage(message *bank.Message, level int) error {
	if message == nil {
		return bank.ErrInvalidMessage
	}
	if len(message.Data) == 0 {
		return bank.ErrNoData
	}

	if message.IsCompressed() {
		// NOOP
		return nil
	}

	if message.IsEncrypted() {
		return ErrOperationNotPermited
	}

	data, err := Compress(message.Data, level)
	if err != nil {
		return err
	}
	message.Data = data
	message.SetCompressed(true)

	return nil
}

func DecompressMessage(message *bank.Message) error {
	if message == nil {
		return bank.ErrInvalidMessage
	}
	if len(message.Data) == 0 {
		return bank.ErrNoData
	}

	if !message.IsCompressed() {
		// NOOP
		return nil
	}

	data, err := Decompress(message.Data)
	if err != nil {
		return err
	}
	message.Data = data
	message.SetCompressed(false)

	return nil
}
