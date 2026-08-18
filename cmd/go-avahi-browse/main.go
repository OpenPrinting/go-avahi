// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// go-avahi-browse utility
//
//go:build linux || freebsd

package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/OpenPrinting/go-avahi"
)

func die(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}

func main() {
	// Obtain defaults from Avahi
	clnt, err := avahi.NewClient(0)
	if err != nil {
		die("Avahi error: %s", err)
	}

	defaultDomain := clnt.GetDomainName()
	clnt.Close()

	// Parse options
	var domain string
	var noaddr bool
	var notxt bool
	var ip4, ip6 bool
	var timeout float64

	flag.CommandLine.SetOutput(os.Stdout)

	flag.StringVar(&domain, "d", defaultDomain, "The domain to browse in")
	flag.BoolVar(&noaddr, "noaddr", false, "Don't resolve service address")
	flag.BoolVar(&notxt, "notxt", false, "Don't resolve TXT record")
	flag.BoolVar(&ip4, "4", false, "Use IPv4")
	flag.BoolVar(&ip6, "6", false, "Use IPv6")
	flag.Float64Var(&timeout, "T",
		float64(avahi.DefaultTimeout)/float64(time.Second),
		"Timeout to browse, in seconds")

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	flag.Parse()

	// Check options
	if timeout < 0 {
		fmt.Fprintf(flag.CommandLine.Output(),
			"invalid value \"%g\" for flag -T: must be non-negative\n",
			timeout)
		flag.Usage()
		os.Exit(2)
	}

	// Prepare lookup flags
	lookupFlags := avahi.LookupFlags(0)
	if noaddr {
		lookupFlags |= avahi.LookupNoAddress
	}

	if notxt {
		lookupFlags |= avahi.LookupNoTXT
	}

	// Prepare protocols
	proto := avahi.ProtocolUnspec
	switch {
	case ip4 && !ip6:
		proto = avahi.ProtocolIP4
	case !ip4 && ip6:
		proto = avahi.ProtocolIP6
	}

	// Call SimpleServiceResolver
	services, err := avahi.SimpleServiceResolver(
		context.Background(),
		avahi.IfIndexUnspec,
		proto,
		flag.Args(),
		domain,
		lookupFlags,
		time.Duration(float64(time.Second)*timeout),
	)

	if err != nil {
		die("Avahi error: %s", err)
	}

	for i, service := range services {
		if i > 0 {
			fmt.Println()
		}

		ifname := fmt.Sprintf("%d", service.IfIdx)
		if iface, err := net.InterfaceByIndex(int(service.IfIdx)); err == nil {
			ifname = fmt.Sprintf("%q", iface.Name)
		}

		fmt.Printf("+ interface %s, domain: %s\n",
			ifname, service.Domain)

		fmt.Printf("  svctype:   %q\n", service.SvcType)
		fmt.Printf("  subtypes:  %q\n", service.SvcSubTypes)
		fmt.Printf("  flags:     %s\n", service.Flags)
		fmt.Printf("  instance:  %q\n", service.InstanceName)
		fmt.Printf("  hostname:  %q\n",
			strings.Join(service.Hostname, ", "))

		var endpoints []string
		for _, endpoint := range service.Endpoints {
			endpoints = append(endpoints, endpoint.String())
		}

		s := "none"
		if endpoints != nil {
			s = strings.Join(endpoints, ", ")
		}

		fmt.Printf("  endpoints: %s\n", s)

		s = "none"
		if len(service.Txt) > 0 {
			s = fmt.Sprintf("%q", service.Txt)
		}

		fmt.Printf("  TXT:       %s\n", s)
	}
}
