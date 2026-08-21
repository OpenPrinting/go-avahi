package avahi

// #include <stdlib.h>
// #include <avahi-common/alternative.h>
import "C"
import "unsafe"

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
