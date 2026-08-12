// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Simplified synchronous API
//
//go:build linux || freebsd

package avahi

import (
	"context"
	"net/netip"
	"sort"
	"time"
)

const (
	// DefaultTimeout is the default timeout for [SimpleServiceResolver]
	// when the timeout parameter is 0. It is set to 2.5s for a balanced
	// approach between responsiveness and discovery completeness.
	//
	// Consider using:
	//   - 2.5s (DefaultTimeoutInteractive): For interactive applications where user
	//     experience is priority and missing rare services is acceptable
	//   - 5.0s (DefaultTimeoutAccurate): For administrative tools or when
	//     service discovery completeness is critical
	DefaultTimeout = 2500 * time.Millisecond

	// DefaultTimeoutInteractive is provided for explicit selection of the
	// responsive discovery mode. See DefaultTimeout documentation for details.
	DefaultTimeoutInteractive = 2500 * time.Millisecond

	// DefaultTimeoutAccurate is provided as a convenience for users who
	// need more complete service discovery at the cost of longer wait times.
	// See DefaultTimeout documentation for usage guidance.
	DefaultTimeoutAccurate = 5000 * time.Millisecond
)

// Service is the service description, returned by the [SimpleServiceResolver]
// function.
type Service struct {
	IfIdx        IfIndex           // Network interface index
	Flags        LookupResultFlags // Lookup flags
	SvcType      string            // Service type
	InstanceName string            // Service instance name
	Domain       string            // Service domain
	Hostname     []string          // Service hostname, typically just a single entry
	Endpoints    []netip.AddrPort  // Service endpoints (host:port)
	Txt          []string          // TXT record ("key=value"...)
}

// SimpleServiceResolver provides a simple synchronous API for discovering and
// resolving network services using mDNS/DNS-SD.
//
// It combines the functionality of ServiceBrowser and ServiceResolver:
//   - Discovers services on the network using the specified network interface
//     (ifidx) and IP protocol (proto)
//   - Resolves each discovered service's hostname, IP address(es), and TXT
//     records
//
// Parameters:
//   - ctx: Context for cancellation and deadline control. If the context has
//     a deadline, the minimum of ctx deadline and timeout will be used.
//   - ifidx: Network interface index to use for discovery. Use [IfIndexUnspec]
//     for all available interfaces.
//   - proto: IP protocol to use (IPv4, IPv6, or both).
//   - svctypes: Service types to discover (e.g., []string{"_http._tcp",
//     "_ssh._tcp"}).  All types are discovered in parallel.
//   - domain: The domain to search for services. Use "" for the default domain
//     (typically "local" based on avahi-daemon configuration).
//   - flags: Lookup flags that control the discovery behavior (see
//     [LookupFlags] documentation for details).
//   - timeout: Maximum duration for the entire discovery process. A value of 0
//     defaults to DefaultTimeout. Note that timeout expiration is not an error;
//     the function returns all services discovered within the time limit.
//
// Returns:
//   - []Service: A slice of discovered and resolved services, each containing
//     service name, hostname, IP addresses, port, and TXT records.
//   - error: Any error that occurred during discovery or resolution. Timeout
//     expiration does not return an error.
func SimpleServiceResolver(
	ctx context.Context,
	ifidx IfIndex,
	proto Protocol,
	svctypes []string,
	domain string,
	flags LookupFlags,
	timeout time.Duration) ([]*Service, error) {

	// If nothing requested, return immediately
	if len(svctypes) == 0 {
		// No Servers, no error
		return nil, nil
	}

	// We need a context with timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	ctx, calcelFunc := context.WithTimeout(ctx, timeout)
	defer calcelFunc()

	// We need Client and Poller
	clnt, err := NewClientWait(ctx, ClientLoopbackWorkarounds)
	if err != nil {
		return nil, err
	}

	defer clnt.Close()

	// We need a poller
	poller := NewPoller()
	poller.AddClient(clnt)

	// Deduplicate svctypes
	svcTypesSeen := make(map[string]struct{}, len(svctypes))
	svcTypesLen := 0

	for _, svctype := range svctypes {
		if _, dup := svcTypesSeen[svctype]; !dup {
			svcTypesSeen[svctype] = struct{}{}
			svctypes[svcTypesLen] = svctype
			svcTypesLen++
		}
	}

	svctypes = svctypes[:svcTypesLen]

	// We a ServiceBrowser per each combination of protocol and service type
	var protos []Protocol
	if proto == ProtocolUnspec {
		protos = []Protocol{ProtocolIP4, ProtocolIP6}
	} else {
		protos = []Protocol{proto}
	}

	cntServiceBrowsers := 0
	for _, svctype := range svctypes {
		for _, proto := range protos {
			browser, err := NewServiceBrowser(
				clnt,
				ifidx,
				proto,
				svctype,
				domain,
				flags&(LookupUseWideArea|LookupUseMulticast))

			if err == ErrNoNetwork {
				// If some protocol is not available, just ignore for now
				continue
			}

			if err != nil {
				return nil, err
			}

			defer browser.Close()
			poller.AddServiceBrowser(browser)
			cntServiceBrowsers++
		}
	}

	// At least one ServiceBrowser must be present
	// If not, this is probably because the requested IPv4/IPv6 protocol
	// was not available. So return ErrNoNetwork at this case.
	if cntServiceBrowsers == 0 {
		return nil, ErrNoNetwork
	}

	// Maps of already discovered things
	type serviceID struct {
		IfIdx        IfIndex
		SvcType      string
		InstanceName string
	}

	discovered := make(map[serviceID]*Service)
	resolversMap := make(map[serviceID][]*ServiceResolver)

	// Now poll for events
	for {
		evnt, err := poller.Poll(ctx)
		if err == context.DeadlineExceeded {
			break
		}

		switch evnt := evnt.(type) {
		case *ServiceBrowserEvent:
			// Only BrowserNew and BrowserRemove events are informative
			if evnt.Event != BrowserNew && evnt.Event != BrowserRemove {
				break
			}

			// Lookup service within already discovered.
			id := serviceID{
				IfIdx:        evnt.IfIdx,
				SvcType:      evnt.SvcType,
				InstanceName: evnt.InstanceName,
			}

			// Handle BrowserRemove event
			if evnt.Event == BrowserRemove {
				for _, resolver := range resolversMap[id] {
					resolver.Close()
				}
				delete(discovered, id)
				delete(resolversMap, id)
				break
			}

			// Add or update the Service structure
			service := discovered[id]
			if service == nil {
				service = &Service{
					IfIdx:        evnt.IfIdx,
					SvcType:      evnt.SvcType,
					InstanceName: evnt.InstanceName,
					Domain:       evnt.Domain,
				}

				discovered[id] = service
			}

			service.Flags |= evnt.Flags

			// Create ServiceResolver unless we already have one
			if resolvers := resolversMap[id]; resolvers == nil {
				for _, proto := range protos {
					for _, addrproto := range protos {
						resolver, err := NewServiceResolver(
							clnt,
							evnt.IfIdx,
							proto,
							evnt.InstanceName,
							evnt.SvcType,
							evnt.Domain,
							addrproto,
							flags,
						)

						if err == ErrNoNetwork {
							continue
						}

						if err != nil {
							return nil, err
						}

						defer resolver.Close()
						resolvers = append(resolvers, resolver)
						poller.AddServiceResolver(resolver)
					}
				}

				resolversMap[id] = resolvers
			}

		case *ServiceResolverEvent:
			// Only ResolverFound event is informative
			if evnt.Event != ResolverFound {
				break
			}

			// Lookup the service
			id := serviceID{
				IfIdx:        evnt.IfIdx,
				SvcType:      evnt.SvcType,
				InstanceName: evnt.InstanceName,
			}

			service := discovered[id]
			if service == nil {
				// Just a spurious event (which is very unlikely
				// to happen, but just in case...)
				//
				// Ignore it
				break
			}

			// Update the service
			service.Hostname = appendUnique(service.Hostname, evnt.Hostname)
			if evnt.Port != 0 && evnt.Addr.IsValid() {
				service.Endpoints = appendUnique(service.Endpoints,
					netip.AddrPortFrom(evnt.Addr, evnt.Port))
			}
			service.Txt = appendUnique(service.Txt, evnt.Txt...)
		}
	}

	// Now convert map of discovered services into the slice
	var services []*Service
	for _, service := range discovered {
		if service.Hostname == nil {
			// No answer from ServiceResolver.
			// Skip the Service.
			continue
		}

		services = append(services, service)
	}

	// And sort services, just for reproducibility
	sort.Slice(services, func(i, j int) bool {
		s1 := services[i]
		s2 := services[j]

		switch {
		case s1.IfIdx != s2.IfIdx:
			return s1.IfIdx < s2.IfIdx
		case s1.SvcType != s2.SvcType:
			return s1.SvcType < s2.SvcType
		case s1.InstanceName != s2.InstanceName:
			return s1.InstanceName < s2.InstanceName
		}

		return false
	})

	// Sort Endpoints, just for reproducibility
	for _, service := range services {
		sort.Slice(service.Endpoints, func(i, j int) bool {
			e1 := service.Endpoints[i]
			e2 := service.Endpoints[j]

			cmp := e1.Addr().Compare(e2.Addr())
			if cmp != 0 {
				return cmp < 0
			}

			return e1.Port() < e2.Port()
		})
	}

	return services, nil
}
