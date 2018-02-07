// Copyright (c) 2018, Boise State University All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

// defaultKeySize is the default number of bytes to user for the cookie value
const defaultKeySize = 32

// newKey generates a random value for a cookie
func (m *Manager) newKey() string {
	x := make([]byte, m.keySize)
	n, err := rand.Read(x)
	if err != nil || n != m.keySize {
		panic("rand.Read failed")
	}
	return fmt.Sprintf("%x", x)
}

// Manager creates and manages cookies
type Manager struct {
	mu sync.Mutex
	// m is a map of sessions, with a 32 bit
	m map[string]*session
	// key is the name of the cookie that stores the session id
	name string
	// keySize is the length in bytes of the key, randomly generated
	keySize int
}

// NewManager creates a safe to use Manager by initializing the map
func NewManager(name string) *Manager {
	return &Manager{
		m:       map[string]*session{},
		name:    name,
		keySize: defaultKeySize,
	}
}

var (
	// ErrInvalidSession indicates that there is no named session in the store
	// with the cookie value
	ErrInvalidSession = errors.New("invalid session")
	// ErrInvalidKey indicates there is no value in the underlying session that
	// matches the key.
	ErrInvalidKey = errors.New("invalid key")
)

// Get reads a cookie from a request, queries the session manager and returns
// the value if available.  If the session is invalid ErrInvalidSession is
// returnted, if the key is invalid, ErrInvalidKey is returned.
func (m *Manager) Get(r *http.Request, key string) (string, error) {
	c, err := r.Cookie(m.name)
	if err != nil {
		return "", err
	}
	m.mu.Lock()
	s, ok := m.m[c.Value]
	m.mu.Unlock()
	if !ok {
		return "", ErrInvalidSession
	}
	v, ok := s.get(key)
	if !ok {
		return "", ErrInvalidKey
	}
	return v, nil
}

// Set creates a value in an underlying session if it exists.  If it doesn't, a
// new session is created.
func (m *Manager) Set(w http.ResponseWriter, r *http.Request, key, val string) error {
	var s *session
	var nk string
	var ok bool
	c, err := r.Cookie(m.name)
	m.mu.Lock()
	if err != nil {
		s = newSession()
		nk = m.newKey()
		m.m[nk] = s
		ok = true
	} else {
		nk = c.Value
		s, ok = m.m[nk]
	}
	m.mu.Unlock()
	if !ok {
		return ErrInvalidSession
	}
	s.set(key, val)
	http.SetCookie(w, &http.Cookie{
		Name:     m.name,
		Value:    nk,
		MaxAge:   2419200,
		HttpOnly: true,
	})
	return nil
}
