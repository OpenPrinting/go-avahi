// CGo binding for Avahi
//
// Copyright (C) 2025 by Prashant Andoriya
// See LICENSE for license terms and conditions
//
// Fuzz tests for PublishFlags string formatting
//
//go:build linux || freebsd

package avahi

import (
	"strings"
	"testing"
)

// FuzzPublishFlagsString fuzzes the PublishFlags.String method with valid, mixed, and adversarial flag values.
func FuzzPublishFlagsString(f *testing.F) {
	// Valid flags
	// Seed corpus with valid single flag values
	f.Add(int(PublishUnique))
	f.Add(int(PublishNoProbe))
	f.Add(int(PublishNoAnnounce))
	f.Add(int(PublishAllowMultiple))
	f.Add(int(PublishNoReverse))
	f.Add(int(PublishNoCookie))
	f.Add(int(PublishUpdate))
	f.Add(int(PublishUseWideArea))
	f.Add(int(PublishUseMulticast))
	f.Add(int(0)) // no flags

	// Mixed / combined flags
	f.Add(int(PublishUnique | PublishNoProbe | PublishUpdate))
	f.Add(int(PublishUseWideArea | PublishUseMulticast))

	// Invalid / adversarial values to test robustness
	f.Add(-1)
	f.Add(1 << 30)
	f.Add(0xffffffff)

	f.Fuzz(func(t *testing.T, v int) {
		flags := PublishFlags(v)
		// String() must never panic for any integer input
		s := flags.String()

		// Must never panic and must not contain malformed separators
		// Output should never contain malformed separators such as double commas
		if strings.Contains(s, ",,") {
			t.Fatalf("PublishFlags.String() returned malformed string: %q", s)
		}
	})
}
