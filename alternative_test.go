//go:build linux || freebsd

package avahi

import "testing"

func TestAlternativeHostname(t *testing.T) {
	hostname := "test-hostname"

	for i := 0; i < 10; i++ {
		alt, err := AlternativeHostname(hostname)
		if err != nil {
			t.Fatalf("AlternativeHostname failed: %v", err)
		}

		t.Logf("Alternative hostname: %s", alt)

		if alt == hostname {
			t.Fatalf("AlternativeHostname did not return an alternative name")
		}

		hostname = alt
	}
}

func TestAlternativeServiceName(t *testing.T) {
	serviceName := "test-service"

	for i := 0; i < 10; i++ {
		alt, err := AlternativeServiceName(serviceName)
		if err != nil {
			t.Fatalf("AlternativeServiceName failed: %v", err)
		}

		t.Logf("Alternative service name: %s", alt)

		if alt == serviceName {
			t.Fatalf("AlternativeServiceName did not return an alternative name")
		}

		serviceName = alt
	}
}
