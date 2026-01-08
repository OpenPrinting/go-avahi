// CGo binding for Avahi
//
// Copyright (C) 2025 by Prashant Andoriya
// See LICENSE for license terms and conditions
//
// Fuzz tests for Protocol string conversion
//
//go:build linux || freebsd

package avahi

import (
	"strings"
	"testing"
)

// FuzzProtocolString fuzzes the Protocol.String method
func FuzzProtocolString(f *testing.F) {
	// Valid protocol values
	f.Add(int(ProtocolIP4))
	f.Add(int(ProtocolIP6))
	f.Add(int(ProtocolUnspec))

	// Invalid / adversarial values
	f.Add(0)
	f.Add(-1)
	f.Add(1 << 29)
	f.Add(0xffffffff)

	f.Fuzz(func(t *testing.T, v int) {
		proto := Protocol(v)
		s := proto.String()

		// Must never panic and must always return a non empty string
		if s == "" {
			t.Fatalf("Protocol.String() returned empty string for value %d", v)
		}

		// For unknown values, String() should indicate UNKNOWN
		if v != int(ProtocolIP4) && v != int(ProtocolIP6) && v != int(ProtocolUnspec) {
			if !strings.HasPrefix(s, "UNKNOWN") {
				t.Fatalf("unexpected string for unknown protocol %d: %q", v, s)
			}
		}
	})
}
