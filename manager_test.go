// Copyright (c) 2018, Boise State University All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestManagerGetRequest(t *testing.T) {
	sm := NewManager("foo")
	r, _ := http.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  "foo",
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

func TestManagerSet(t *testing.T) {
	sm := NewManager("mine")
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
	if ck[0] != "mine" {
		t.Errorf("get/set failed, got: %s, want: %s", ck[0], "mine")
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
