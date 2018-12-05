// Copyright (c) 2018, Boise State University All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	// defaultKeySize is the default number of bytes to user for the cookie value
	defaultKeySize = 32

	// debug prints debugging info to log
	debug = false
)

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
	// maxAge is the lifetime of the session, the sessions are cleared if they
	// aren't accessed within this lifetime.
	maxAge time.Duration
	// c is a channel to kill the flushing of expired sessions
	c chan struct{}
}

// NewManager creates a safe to use Manager by initializing the map
func NewManager(name string) *Manager {
	m := &Manager{
		m:       map[string]*session{},
		name:    name,
		keySize: defaultKeySize,
		// TODO(kyle): allow this to be set
		maxAge: 2419200 * time.Second,
		c:      make(chan struct{}),
	}
	go m.StartExpunge()
	return m
}

var (
	// ErrInvalidSession indicates that there is no named session in the store
	// with the cookie value
	ErrInvalidSession = errors.New("invalid session")
	// ErrInvalidKey indicates there is no value in the underlying session that
	// matches the key.
	ErrInvalidKey = errors.New("invalid key")
)

func (m *Manager) StartExpunge() {
	t := time.NewTicker(m.maxAge / 4)
	for {
		select {
		case <-t.C:
			m.mu.Lock()
			for k, v := range m.m {
				if time.Now().After(v.accessed.Add(m.maxAge)) {
					if debug {
						log.Printf("expunging session %s: %#v", k, v)
					}
					delete(m.m, k)
				}
			}
			m.mu.Unlock()
		case <-m.c:
			return
		}
	}
}

func (m *Manager) StopExpunge() {
	m.c <- struct{}{}
}

// Get reads a cookie from a request, queries the session manager and returns
// the value if available.  If the session is invalid ErrInvalidSession is
// returnted, if the key is invalid, ErrInvalidKey is returned.
func (m *Manager) Get(r *http.Request, key string) (string, error) {
	if debug {
		log.Printf("attempting to fetch %s", key)
		log.Printf("state: %s", m.state())
	}
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
	s.accessed = time.Now()
	return v, nil
}

// Set creates a value in an underlying session if it exists.  If it doesn't, a
// new session is created.
func (m *Manager) Set(w http.ResponseWriter, r *http.Request, key, val string) error {
	var s *session
	var nk string
	var ok bool
	if debug {
		log.Printf("setting %s to %s", key, val)
		log.Printf("state before: %s", m.state())
	}
	c, err := r.Cookie(m.name)
	if err != nil {
		s = newSession()
		nk = m.newKey()
	} else {
		nk = c.Value
	}
	s, ok = m.m[nk]
	if !ok {
		s = newSession()
	}
	m.m[nk] = s
	s.set(key, val)
	c = &http.Cookie{
		Name:     m.name,
		Value:    nk,
		MaxAge:   int(m.maxAge.Seconds()),
		HttpOnly: true,
	}
	http.SetCookie(w, c)
	if debug {
		log.Printf("state after: %s", m.state())
	}
	return nil
}

// state dumps the current state of the manager and sessions to the stdout log
func (m *Manager) state() string {
	s := fmt.Sprintf("name: %s\n", m.name)
	s += fmt.Sprintf("max-age: %d\n", m.maxAge)
	s += "sessions:\n"
	for k, v := range m.m {
		s += fmt.Sprintf("\tkey: %s\n", k)
		s += "\tkey\tvalue\n"
		for kk, vv := range v.m {
			s += fmt.Sprintf("\t%s\t%s\n", kk, vv)
		}
	}
	return s
}
