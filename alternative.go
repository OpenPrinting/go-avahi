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

// AlternativeHostname finds an alternative for the specified host name:
//
//	"foo" -> "foo-2"
//	"foo-2" -> "foo-3"
func AlternativeHostname(hostname string) (string, error) {
	chostname := C.CString(hostname)
	defer C.free(unsafe.Pointer(chostname))

	alt := C.avahi_alternative_host_name(chostname)
	if alt == nil {
		return "", ErrNoMemory
	}

	defer C.free(unsafe.Pointer(alt))
	return C.GoString(alt), nil
}

// AlternativeServiceName finds an alternative for the specified service name:
//
//	"Foo" -> "Foo #2"
//	"Foo #2" -> "Foo #3"
func AlternativeServiceName(name string) (string, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	alt := C.avahi_alternative_service_name(cname)
	if alt == nil {
		return "", ErrNoMemory
	}

	defer C.free(unsafe.Pointer(alt))
	return C.GoString(alt), nil
}
