// CGo binding for Avahi
//
// Copyright (C) 2025 by Prashant Andoriya
// See LICENSE for license terms and conditions
//
// Unit tests for Protocol string conversion
//
//go:build linux || freebsd

package avahi

import "testing"

// TestProtocolString verifies string output for known protocol values
func TestProtocolString(t *testing.T) {
	tests := []struct {
		proto Protocol
		want  string
	}{
		{ProtocolIP4, "ip4"},
		{ProtocolIP6, "ip6"},
		{ProtocolUnspec, "unspec"},
	}

	for _, tt := range tests {
		if got := tt.proto.String(); got != tt.want {
			t.Fatalf("proto=%v: got %q, want %q", tt.proto, got, tt.want)
		}
	}
}

// TestProtocolStringUnknown verifies handling of unknown protocol values
func TestProtocolStringUnknown(t *testing.T) {
	p := Protocol(12345)
	s := p.String()

	if s == "" {
		t.Fatalf("Protocol.String() returned empty string for unknown protocol")
	}

	if s[:7] != "UNKNOWN" {
		t.Fatalf("unexpected string for unknown protocol: %q", s)
	}
}
