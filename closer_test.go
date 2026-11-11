// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Closers tests
//
//go:build linux || freebsd

package avahi

import "testing"

type mockCloser struct {
	closed bool
}

func (m *mockCloser) Close() {
	m.closed = true
}

func TestClosers(t *testing.T) {
	t.Run("Lifecycle", func(t *testing.T) {
		var set closers
		set.init()

		c1 := &mockCloser{}
		c2 := &mockCloser{}

		set.add(c1)
		set.add(c2)

		if len(set) != 2 {
			t.Errorf("expected 2 closers, got %d", len(set))
		}

		set.del(c1)
		if len(set) != 1 {
			t.Errorf("expected 1 closer after deletion, got %d", len(set))
		}

		set.close()
		if !c2.closed {
			t.Error("c2 should have been closed")
		}
		if c1.closed {
			t.Error("c1 should not have been closed (it was deleted from the set)")
		}
	})

	t.Run("EmptySet", func(t *testing.T) {
		var set closers
		set.init()
		set.close() // Should not panic
	})
}
