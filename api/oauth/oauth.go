package oauth

import (
	"errors"

	"github.com/joho/godotenv"
)

var (
	ErrInvalidOAuthKeys = errors.New("Invalid OAuth keys file")
)

type Options struct {
	Keys string
}

func Init(options Options) error {
	err := godotenv.Overload(options.Keys)
	if err != nil {
		return ErrInvalidOAuthKeys
	}
	return nil
}
