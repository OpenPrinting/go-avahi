// CGo binding for Avahi
//
// Copyright (C) 2025 by Prashant Andoriya
// See LICENSE for license terms and conditions
//
// Fuzz tests for eventqueue robustness
//
//go:build linux || freebsd

package avahi

import "testing"

// FuzzEventQueueOperations fuzzes push and close operations on eventqueue
func FuzzEventQueueOperations(f *testing.F) {
	// Seed inputs
	f.Add(0)
	f.Add(1)
	f.Add(10)
	f.Add(100)

	f.Fuzz(func(t *testing.T, n int) {
		if n < 0 {
			return
		}
		if n > 1000 {
			n = 1000
		}

		var q eventqueue[int]
		q.init()

		// Push values
		for i := 0; i < n; i++ {
			q.Push(i)
		}

		// Drain some values (best effort)
		for i := 0; i < n/2; i++ {
			select {
			case <-q.Chan():
			default:
			}
		}
		q.Close()
	})
}
