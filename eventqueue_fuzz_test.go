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

// FuzzEventQueueOperations fuzzes combinations of Push, read,
// and Close operations to validate that eventqueue:
//
//   - never panics
//   - never deadlocks
//   - always allows Close() to complete safely
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

		// Push n values into the queue
		for i := 0; i < n; i++ {
			q.Push(i)
		}

		// Drain some values on a best effort basis.
		// Partial reads are intentional and valid.
		for i := 0; i < n/2; i++ {
			select {
			case <-q.Chan():
			default:
			}
		}
		// Close must always complete without panic or deadlock
		q.Close()
	})
}
