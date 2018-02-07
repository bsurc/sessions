// Copyright (c) 2018, Boise State University All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import "sync"

// session is a safe to use session for web apps
type session struct {
	// Embed Mutex for locking and unlocking the session, namely the map
	sync.Mutex
	// m stores session data
	m map[string]string
}

// newSession initializes the map of a session and returns a valid new session
func newSession() *session {
	return &session{
		m: map[string]string{},
	}
}

// get retrieves the value paired with key in map m while guarding
func (s *session) get(key string) (string, bool) {
	s.Lock()
	x, ok := s.m[key]
	s.Unlock()
	return x, ok
}

// set sets the value paired with key in map m while guarding
func (s *session) set(key, val string) error {
	s.Lock()
	s.m[key] = val
	s.Unlock()
	return nil
}
