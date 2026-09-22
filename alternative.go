// CGo binding for Avahi
//
// Copyright (C) 2026 and up by Gianluca Altomani (altomanigianluca@gmail.com)
// See LICENSE for license terms and conditions
//
// Alternative names
//
//go:build linux || freebsd

package avahi

// #include <stdlib.h>
// #include <avahi-common/alternative.h>
import "C"
import "unsafe"

// AlternativeHostname suggests an alternative name for the specified host name
// by incrementing its numeric suffix:
//
//	"foo"   -> "foo-2"
//	"foo-2" -> "foo-3"
//
// This function is intended for automatic collision resolution when
// publishing a hostname via DNS-SD. It returns an error if the provided
// hostname has an invalid syntax.
func AlternativeHostname(hostname string) (string, error) {
	chostname := C.CString(hostname)
	defer C.free(unsafe.Pointer(chostname))

	alt := C.avahi_alternative_host_name(chostname)
	if alt == nil {
		// Note: Returning ErrInvalidHostName is a best-effort guess
		// because avahi_alternative_host_name does not return a
		// specific error code. If it fails, it is most likely due to
		// an invalid hostname syntax.
		return "", ErrInvalidHostName
	}

	defer C.free(unsafe.Pointer(alt))
	return C.GoString(alt), nil
}

// AlternativeServiceName suggests an alternative name for the specified
// service name by incrementing its numeric suffix:
//
//	"Foo"    -> "Foo #2"
//	"Foo #2" -> "Foo #3"
//
// This function is intended for automatic collision resolution when
// publishing a service via DNS-SD. It returns an error if the provided
// service name has an invalid syntax.
func AlternativeServiceName(name string) (string, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	alt := C.avahi_alternative_service_name(cname)
	if alt == nil {
		// Note: Returning ErrInvalidServiceName is a best-effort guess
		// because avahi_alternative_service_name does not return a
		// specific error code. If it fails, it is most likely due to
		// an invalid service name syntax.
		return "", ErrInvalidServiceName
	}

	defer C.free(unsafe.Pointer(alt))
	return C.GoString(alt), nil
}
