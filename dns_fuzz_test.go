//go:build linux || freebsd

package avahi

import (
	"net/netip"
	"testing"
)

func FuzzDNSDecodeA(f *testing.F) {
	// Valid IPv4
	f.Add([]byte{127, 0, 0, 1})
	f.Add([]byte{8, 8, 8, 8})

	// Invalid sizes
	f.Add([]byte{})
	f.Add([]byte{1})
	f.Add([]byte{1, 2, 3})
	f.Add([]byte{1, 2, 3, 4, 5})

	// Random bytes
	f.Add([]byte{255, 255, 255, 255})

	f.Fuzz(func(t *testing.T, data []byte) {
		addr := DNSDecodeA(data)

		if addr.IsValid() && addr.Is6() {
			t.Fatalf("DNSDecodeA returned IPv6 address: %v", addr)
		}
	})
}

func FuzzDNSDecodeAAAA(f *testing.F) {
	// Valid IPv6 (loopback)
	f.Add(netip.IPv6Loopback().AsSlice())

	// Invalid sizes
	f.Add([]byte{})
	f.Add([]byte{1, 2, 3})
	f.Add(make([]byte, 15))
	f.Add(make([]byte, 17))

	f.Fuzz(func(t *testing.T, data []byte) {
		addr := DNSDecodeAAAA(data)

		if addr.IsValid() && !addr.Is6() {
			t.Fatalf("DNSDecodeAAAA returned non-IPv6 address: %v", addr)
		}
	})
}

func FuzzDNSDecodeTXT(f *testing.F) {
	// Valid TXT records
	f.Add([]byte{3, 'f', 'o', 'o'})
	f.Add([]byte{3, 'f', 'o', 'o', 3, 'b', 'a', 'r'})

	// Empty string
	f.Add([]byte{0})

	// Length exceeds data
	f.Add([]byte{10, 'a'})

	// Nested garbage
	f.Add([]byte{255, 255, 255})

	// Random bytes
	f.Add([]byte{1, 0, 255, 10, 20, 30})

	f.Fuzz(func(t *testing.T, data []byte) {
		txt := DNSDecodeTXT(data)

		if txt != nil {
			for _, s := range txt {
				if len(s) == 0 {
					t.Fatalf("DNSDecodeTXT returned empty string")
				}
			}
		}
	})
}
