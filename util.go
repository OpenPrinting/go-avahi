// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Miscellaneous utility functions
//
//go:build linux || freebsd

package avahi

import "slices"

// appendUnique appends elements to the end of slice, ignoring values already
// present in the slice.
func appendUnique[T comparable](slice []T, elems ...T) []T {
	for _, elem := range elems {
		if !slices.Contains(slice, elem) {
			slice = append(slice, elem)
		}
	}

	return slice
}
