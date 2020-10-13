package sessions

import (
	"errors"
)

var (
	ErrInternalError = errors.New("InternalError")

	ErrInvalidDuration   = errors.New("Invalid Duration")
	ErrInvalidUserID     = errors.New("Invalid UserID")
	ErrInvalidRemoteAddr = errors.New("Invalid RemoteAddr")
	ErrRemoteAddrChanged = errors.New("RemoteAddr Changed")
	ErrInvalidSessionID  = errors.New("Invalid SessionID")
	ErrSessionExpired    = errors.New("Session Expired")
	ErrEncode            = errors.New("Encode Error")
	ErrDecode            = errors.New("Decode Error")
	ErrCache             = errors.New("Cache Error")

	ErrInvalidCrendential    = errors.New("InvalidCredentials")
	ErrMissingCookie         = errors.New("MissingCookie")
	ErrInvalidCookie         = errors.New("ErrInvalidCookie")
	ErrSessionCreationFailed = errors.New("SessionCreationFailed")
	ErrTooManyOpenSession    = errors.New("TooManyOpenSession")
	ErrSessionClose          = errors.New("SessionCloseFailed")
)
