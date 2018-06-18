// Copyright (c) 2018, Boise State University All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const managerKey = "buster"

func TestManagerGetRequest(t *testing.T) {
	sm := NewManager(managerKey)
	r, _ := http.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  managerKey,
		Value: "bar",
	})
	s := newSession()
	sm.m["bar"] = s
	s.set("boo", "baz")
	x, err := sm.Get(r, "boo")
	if err != nil {
		t.Fatal(err)
	}
	if x != "baz" {
		t.Errorf("get failed, got: %s, want: %s", x, "baz")
	}
}

func TestManagerGetNothing(t *testing.T) {
	sm := NewManager(managerKey)
	req, _ := http.NewRequest("GET", "/", nil)
	_, err := sm.Get(req, "")
	if err != http.ErrNoCookie {
		t.Errorf("invalid get, expected %s, got %s", http.ErrNoCookie, err)
	}
}

func TestManagerSet(t *testing.T) {
	sm := NewManager(managerKey)
	resp := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sm.Set(w, r, "foo", "bar")
		w.WriteHeader(http.StatusOK)
	})

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	mux.ServeHTTP(resp, req)

	// Make sure the header is set.
	hdr := resp.HeaderMap["Set-Cookie"][0]
	tkns := strings.Split(hdr, ";")
	ck := strings.Split(tkns[0], "=")
	if ck[0] != managerKey {
		t.Errorf("get/set failed, got: %s, want: %s", ck[0], managerKey)
	}

	s, ok := sm.m[ck[1]]
	if !ok {
		t.Error("invalid cookie")
	}
	x, ok := s.get("foo")
	if !ok || x != "bar" {
		t.Errorf("set/get failed, got: %s, want: %s (ok==%t)", x, "bar", ok)
	}
}

func TestInvalidKey(t *testing.T) {
	sm := NewManager(managerKey)
	s := newSession()
	s.set("foo", "bar")
	k := sm.newKey()
	sm.m[k] = s

	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  managerKey,
		Value: k,
	})
	_, err := sm.Get(req, "baz")
	if err != ErrInvalidKey {
		t.Errorf("invalid get, expected %s, got %s", ErrInvalidKey, err)
	}
}

func TestInvalidSession(t *testing.T) {
	sm := NewManager(managerKey)
	s := newSession()
	s.set("foo", "bar")
	k := sm.newKey()
	sm.m[k] = s

	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  managerKey,
		Value: k,
	})

	v, err := sm.Get(req, "foo")
	if err != nil {
		t.Error(err)
	}
	if v != "bar" {
		t.Errorf("set/get failed, got: %s, want: %s", v, "bar")
	}

	nk := sm.newKey()
	req, _ = http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  managerKey,
		Value: nk,
	})

	v, err = sm.Get(req, "foo")
	if err != ErrInvalidSession {
		t.Errorf("invalid get, expected %s, got %s", ErrInvalidSession, err)
	}

	resp := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sm.Set(w, r, "foo", "bar")
		w.WriteHeader(http.StatusOK)
	})
	mux.ServeHTTP(resp, req)
	c := resp.Header().Get("Set-Cookie")
	if c == "" {
		t.Fatal("failed to set cookie")
	}
	if strings.Index(c, nk) < 0 {
		t.Error("failed to set cookie properly")
	}
}

func TestExpunge(t *testing.T) {
	sm := NewManager(managerKey)
	sm.maxAge = time.Second
	s := newSession()
	s.set("foo", "bar")
	k := sm.newKey()
	sm.m[k] = s
	if _, ok := sm.m[k]; !ok {
		t.Error("session should be alive")
	}
	time.Sleep(2 * time.Second)
	if s, ok := sm.m[k]; ok {
		t.Errorf("session %#v should be cleared", s)
	}

}
