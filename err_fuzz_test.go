// CGo binding for Avahi
//
// Copyright (C) 2025 by Prashant Andoriya
// See LICENSE for license terms and conditions
//
// Fuzz tests for Avahi error code string conversion
//
//go:build linux || freebsd

package avahi

import "testing"

// FuzzErrCodeError fuzzes the ErrCode.Error method to verify that:
//
//   - it never panics for arbitrary integer values
//   - it always returns a non empty string
//
// This is especially important because ErrCode.Error() crosses the Go - C boundary and relies on avahi_strerror()
func FuzzErrCodeError(f *testing.F) {
	// Valid error codes
	f.Add(int(NoError))
	f.Add(int(ErrFailure))
	f.Add(int(ErrInvalidHostName))
	f.Add(int(ErrTimeout))
	f.Add(int(ErrInvalidFlags))
	f.Add(int(ErrDNSNXDOMAIN))

	// Invalid / adversarial values
	f.Add(0)
	f.Add(-1)
	f.Add(1 << 29)
	f.Add(0xffffffff)

	f.Fuzz(func(t *testing.T, v int) {
		err := ErrCode(v)
		s := err.Error()

		// Must never panic and must always return a non empty string
		if s == "" {
			t.Fatalf("ErrCode.Error() returned empty string for value %d", v)
		}
	})
}
