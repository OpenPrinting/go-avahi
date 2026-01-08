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

// FuzzPublishFlagsString fuzzes the PublishFlags.String method
func FuzzPublishFlagsString(f *testing.F) {
	// Valid flags
	f.Add(int(PublishUnique))
	f.Add(int(PublishNoProbe))
	f.Add(int(PublishNoAnnounce))
	f.Add(int(PublishAllowMultiple))
	f.Add(int(PublishNoReverse))
	f.Add(int(PublishNoCookie))
	f.Add(int(PublishUpdate))
	f.Add(int(PublishUseWideArea))
	f.Add(int(PublishUseMulticast))
	f.Add(int(0))

	// Mixed / combined flags
	f.Add(int(PublishUnique | PublishNoProbe | PublishUpdate))
	f.Add(int(PublishUseWideArea | PublishUseMulticast))

	// Invalid / adversarial values
	f.Add(-1)
	f.Add(1 << 30)
	f.Add(0xffffffff)

	f.Fuzz(func(t *testing.T, v int) {
		flags := PublishFlags(v)
		s := flags.String()

		// Must never panic and must not contain malformed separators
		if strings.Contains(s, ",,") {
			t.Fatalf("PublishFlags.String() returned malformed string: %q", s)
		}
	})
}
