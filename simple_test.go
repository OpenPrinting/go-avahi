// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Simple API test
//
//go:build linux || freebsd

package avahi

import (
	"bytes"
	"context"
	"fmt"
	"net/netip"
	"sync"
	"testing"
	"time"
)

func TestSimpleAPI(t *testing.T) {
	// Constant parameters
	timeout := 5 * time.Second
	loopback := MustLoopback()
	instancename := randName()
	addrIP4 := netip.MustParseAddrPort("127.0.0.1:80")
	addrIP6 := netip.MustParseAddrPort("[::1]:80")
	addrIP4tls := netip.MustParseAddrPort("127.0.0.1:443")
	addrIP6tls := netip.MustParseAddrPort("[::1]:443")

	// Prepare test data
	services := []*Service{
		{
			IfIdx:        loopback,
			SvcType:      "_ipp._tcp",
			SvcSubTypes:  []string{"_universal._sub._ipp._tcp"},
			InstanceName: instancename,
			Hostnames:    []string{"localhost"},
			Endpoints:    []netip.AddrPort{addrIP4, addrIP6},
			Txt: []string{
				"txtvers=1",
				"pdl=image/pwg-raster",
				"ty=OpenPrinting Test Printer",
			},
		},

		{
			IfIdx:        loopback,
			SvcType:      "_ipps._tcp",
			SvcSubTypes:  []string{"_universal._sub._ipps._tcp"},
			InstanceName: instancename,
			Hostnames:    []string{"localhost"},
			Endpoints:    []netip.AddrPort{addrIP4tls, addrIP6tls},
			Txt: []string{
				"txtvers=1",
				"pdl=image/pwg-raster",
				"ty=OpenPrinting Test Printer",
			},
		},

		{
			IfIdx:        loopback,
			SvcType:      "_printer._tcp",
			InstanceName: instancename,
			Hostnames:    []string{"localhost"},
		},
	}

	sortServices(services)

	// Run SimpleServicePublisher in background
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	var wait sync.WaitGroup
	wait.Add(1)
	go func() {
		SimpleServicePublisher(ctx, ProtocolUnspec, 0, done, services)
		fmt.Println("DONE SimpleServicePublisher")
		wait.Done()
	}()

	defer func() {
		cancel()
		wait.Wait()
	}()

	// Wait until publishing is done
	tm := time.NewTimer(timeout)
	select {
	case <-done:
	case <-tm.C:
		t.Fatalf("SimpleServicePublisher timeout")
	}

	// Create a collision. Note, Avahi considers collision
	// when name is the same, but service data doesn't match,
	// so we need a second copy of the services slice
	collision := make([]*Service, len(services))
	for i := range services {
		svc := *services[i]
		collision[i] = &svc
		collision[i].Endpoints = nil
	}

	err := SimpleServicePublisher(ctx, ProtocolUnspec, 0, nil, collision)
	if err != ErrCollision {
		t.Errorf("SimpleServicePublisher:\n"+
			"error expected: %s\n"+
			"error present:  %v\n", ErrCollision, err)
	}

	// Do resolve
	tmo, cancel2 := context.WithTimeout(context.Background(), timeout)
	defer cancel2()

	types := []string{}
	for _, svc := range services {
		types = append(types, svc.SvcType)
		types = append(types, svc.SvcSubTypes...)
	}

	resolved, err := SimpleServiceResolver(
		tmo,
		loopback,
		ProtocolUnspec,
		types,
		"",
		0,
		1*time.Second,
	)

	if err != nil {
		t.Errorf("SimpleServiceResolver: %v", err)
		return
	}

	// We are only interested in services that match
	// the instancename
	end := 0
	for i := range resolved {
		if resolved[i].InstanceName == instancename {
			resolved[end] = resolved[i]
			end++
		}
	}
	resolved = resolved[:end]

	// Format published services and actually resolved
	var expected, present bytes.Buffer
	for _, svc := range services {
		fmt.Fprintf(&expected, "SvcType:      %s\n", svc.SvcType)
		fmt.Fprintf(&expected, "SvcSubTypes:  %s\n", svc.SvcSubTypes)
		fmt.Fprintf(&expected, "InstanceName: %s\n", svc.InstanceName)
		fmt.Fprintf(&expected, "Hostnames:    %s\n", svc.Hostnames)
		fmt.Fprintf(&expected, "Endpoints:    %s\n", svc.Endpoints)
		fmt.Fprintf(&expected, "\n")
	}

	for _, svc := range resolved {
		fmt.Fprintf(&present, "SvcType:      %s\n", svc.SvcType)
		fmt.Fprintf(&present, "SvcSubTypes:  %s\n", svc.SvcSubTypes)
		fmt.Fprintf(&present, "InstanceName: %s\n", svc.InstanceName)
		fmt.Fprintf(&present, "Hostnames:    %s\n", svc.Hostnames)
		fmt.Fprintf(&present, "Endpoints:    %s\n", svc.Endpoints)
		fmt.Fprintf(&present, "\n")
	}

	if expected.String() != present.String() {
		t.Errorf("Publish/Resolve mismatc:\n"+
			"expected:\n"+
			"%s\n"+
			"present:\n"+
			"%s\n", expected.String(), present.String())
	}
}
