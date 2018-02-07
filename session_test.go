// Copyright (c) 2018, Boise State University All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import "testing"

func TestNewSession(t *testing.T) {
	s := newSession()
	if s == nil {
		t.Fatal("nil session")
	}
}

func TestSessionSetGet(t *testing.T) {
	s := newSession()
	s.set("foo", "bar")
	x, ok := s.get("foo")
	if !ok || x != "bar" {
		t.Errorf("set/get failed, got: %s, want: %s (ok==%t)", x, "bar", ok)
	}
}

func TestSessionGetInvalid(t *testing.T) {
	s := newSession()
	x, ok := s.get("foo")
	if ok || x != "" {
		t.Errorf("get failed, got: %s, want: %s (ok==%t)", x, "", ok)
	}
}
