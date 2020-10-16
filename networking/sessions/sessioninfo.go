// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sessions

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/google/uuid"
)

type SessionID string

const (
	cstInvalidUserID     = uint64(0)
	cstInvalidSessionID  = SessionID("")
	cstInvalidRemoteAddr = ""
)

func NewSessionID() SessionID {
	return SessionID(uuid.New().String())
}

type SessionInfo struct {
	UserID     uint64
	RemoteAddr string
	SessionID  SessionID
	Expiration time.Time
}

func (s *SessionInfo) Expired() bool {
	if s.SessionID == cstInvalidSessionID {
		return true
	}

	now := time.Now().UTC()
	return now.After(s.Expiration)
}

// Encode return bytes from Message. Encoded with gob
func (s *SessionInfo) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)

	err := enc.Encode(s)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Decode return SessionInfo from bytes. Decoded with gob
func (s *SessionInfo) Decode(data []byte) error {
	buffer := bytes.NewReader(data)
	dec := gob.NewDecoder(buffer)

	err := dec.Decode(s)
	if err != nil {
		return err
	}
	return nil
}
