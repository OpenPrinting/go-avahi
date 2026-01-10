// CGo binding for Avahi
//
// Copyright (C) 2025 by Prashant Andoriya
// See LICENSE for license terms and conditions
//
// Unit tests for Avahi error code handling
//
//go:build linux || freebsd

package avahi

import (
	"strings"
	"testing"
)

// Compile time assertion that ErrCode implements the error interface.
// This will fail to compile if ErrCode no longer satisfies error.
var _ error = ErrFailure

// TestErrCodeError verifies that ErrCode.Error() returns a non-empty,
//
// The test intentionally does not assert the exact error message, as
// the underlying string is provided by the Avahi C library and may vary across versions or environments.
func TestErrCodeError(t *testing.T) {
	tests := []ErrCode{
		NoError,
		ErrFailure,
		ErrInvalidHostName,
		ErrInvalidDomainName,
		ErrTimeout,
		ErrInvalidFlags,
		ErrDNSNXDOMAIN,
	}

	for _, ec := range tests {
		s := ec.Error()
		if s == "" {
			t.Fatalf("ErrCode(%d).Error() returned empty string", ec)
		}
		if !strings.HasPrefix(s, "avahi: ") {
			t.Fatalf("unexpected error prefix for ErrCode(%d): %q", ec, s)
		}
	}
}
