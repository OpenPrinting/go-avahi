// CGo binding for Avahi
//
// Copyright (C) 2025 by Prashant Andoriya
// See LICENSE for license terms and conditions
//
// Fuzz tests for Avahi lookup flag string formatting.
//
//go:build linux || freebsd

package avahi

import (
	"strings"
	"testing"
)

// FuzzLookupFlagsString fuzzes the LookupFlags.String method
func FuzzLookupFlagsString(f *testing.F) {
	// Valid combinations
	f.Add(int(LookupUseWideArea))
	f.Add(int(LookupUseMulticast))
	f.Add(int(LookupNoTXT))
	f.Add(int(LookupNoAddress))
	f.Add(int(LookupUseWideArea | LookupNoTXT))
	f.Add(int(0))

	// Invalid / random combinations
	f.Add(-1)
	f.Add(1 << 30)
	f.Add(0xffffffff)

	f.Fuzz(func(t *testing.T, v int) {
		flags := LookupFlags(v)
		s := flags.String()

		if strings.Contains(s, " ") {
			t.Fatalf("LookupFlags.String() returned invalid string: %q", s)
		}
	})
}

// FuzzLookupResultFlagsString fuzzes the LookupResultFlags.String method
func FuzzLookupResultFlagsString(f *testing.F) {
	// Valid flags
	f.Add(int(LookupResultCached))
	f.Add(int(LookupResultWideArea))
	f.Add(int(LookupResultMulticast))
	f.Add(int(LookupResultLocal))
	f.Add(int(LookupResultOurOwn))
	f.Add(int(LookupResultStatic))
	f.Add(int(0))

	// Invalid / random combinations
	f.Add(-1)
	f.Add(1 << 29)
	f.Add(0xffffffff)

	f.Fuzz(func(t *testing.T, v int) {
		flags := LookupResultFlags(v)
		s := flags.String()

		// String() must never panic
		// Ensure no empty tokens like ",,"
		if strings.Contains(s, ",,") {
			t.Fatalf("LookupResultFlags.String() returned malformed string: %q", s)
		}
	})
}
