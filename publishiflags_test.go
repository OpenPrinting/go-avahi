// CGo binding for Avahi
//
// Copyright (C) 2025 by Prashant Andoriya
// See LICENSE for license terms and conditions
//
// Unit tests for PublishFlags string formatting
//
//go:build linux || freebsd

package avahi

import "testing"

// TestPublishFlagsString verifies correct string output for
// individual flags and common flag combinations.
func TestPublishFlagsString(t *testing.T) {
	tests := []struct {
		flags PublishFlags
		want  string
	}{
		// No flags should produce an empty string
		{0, ""},
		// Individual flags
		{PublishUnique, "unique"},
		{PublishNoProbe, "no-probe"},
		{PublishNoAnnounce, "no-announce"},
		{PublishAllowMultiple, "allow-multiple"},
		{PublishNoReverse, "no-reverse"},
		{PublishNoCookie, "no-cookie"},
		{PublishUpdate, "update"},
		{PublishUseWideArea, "use-wan"},
		{PublishUseMulticast, "use-mdns"},
		// Common flag combinations
		{PublishUnique | PublishNoProbe, "unique,no-probe"},
		{PublishUpdate | PublishUseWideArea, "update,use-wan"},
	}

	for _, tt := range tests {
		if got := tt.flags.String(); got != tt.want {
			t.Fatalf("flags=%v: got %q, want %q", tt.flags, got, tt.want)
		}
	}
}
