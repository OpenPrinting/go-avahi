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

// TestErrCodeError verifies that Error() returns a prefixed error string
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

// TestErrCodeImplementsError verifies ErrCode implements the error interface
func TestErrCodeImplementsError(t *testing.T) {
	var err error = ErrFailure
	if err == nil {
		t.Fatalf("ErrCode does not implement error interface")
	}
}
