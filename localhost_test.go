// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Localhost handling tests
//
//go:build linux || freebsd

package avahi

import "testing"

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		hostname string
		expected bool
	}{
		{
			hostname: "localhost",
			expected: true,
		},
		{
			hostname: "localhost.localdomain",
			expected: true,
		},
		{
			hostname: "LOCALHOST",
			expected: true,
		},
		{
			hostname: "LocalHost.LocalDomain",
			expected: true,
		},
		{
			hostname: "example.com",
			expected: false,
		},
		{
			hostname: "127.0.0.1",
			expected: false, // isLocalhost only checks string matches
		},
		{
			hostname: "",
			expected: false,
		},
	}

	for _, test := range tests {
		result := isLocalhost(test.hostname)
		if result != test.expected {
			t.Errorf("isLocalhost(%q): expected %v, got %v", test.hostname, test.expected, result)
		}
	}
}
