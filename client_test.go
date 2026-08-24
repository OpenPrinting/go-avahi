// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi Client tests

//go:build linux || freebsd

package avahi

import (
	"context"
	"testing"
	"time"
)

// newTestClient creates a Client with a timeout, so tests fail instead of
// blocking indefinitely if the Avahi daemon is unavailable.
func newTestClient(t *testing.T) *Client {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clnt, err := NewClientWait(ctx, ClientLoopbackWorkarounds)
	if err != nil {
		t.Fatal(err)
	}

	return clnt
}

// mustPanicWith tests that function f panics with the expected
// reason.
func mustPanicWith(t *testing.T, want string, f func()) {
	t.Helper()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic %q", want)
		}
		if r != want {
			t.Fatalf("unexpected panic: got %q, want %q", r, want)
		}
	}()

	f()
}

// TestClientClose tests that closing a Client more than once is safe.
func TestClientClose(t *testing.T) {
	clnt := newTestClient(t)

	clnt.Close()
	clnt.Close()
}

// TestClientUseAfterClose tests that using a Client after Close panics.
func TestClientUseAfterClose(t *testing.T) {
	clnt := newTestClient(t)
	clnt.Close()

	mustPanicWith(t, "Client used after Client.Close()", func() {
		clnt.GetVersionString()
	})
}

// TestNewChildAfterClientClose tests that creating a child object after
// Client.Close panics.
func TestNewChildAfterClientClose(t *testing.T) {
	clnt := newTestClient(t)
	clnt.Close()

	mustPanicWith(t, "Client used after Client.Close()", func() {
		_, _ = NewEntryGroup(clnt)
	})
}

// TestClientCloseClosesChildren tests that Client.Close closes existing
// child objects.
func TestClientCloseClosesChildren(t *testing.T) {
	clnt := newTestClient(t)

	egrp, err := NewEntryGroup(clnt)
	if err != nil {
		t.Fatal(err)
	}

	clnt.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	evnt, err := egrp.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if evnt != nil {
		t.Fatalf("unexpected EntryGroup event: %v", evnt)
	}
}
