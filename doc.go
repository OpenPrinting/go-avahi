// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Package documentation
//
//go:build linux || freebsd

/*
Package avahi provides a fairly complete CGo binding for Avahi client.

Avahi is the standard implementation of Multicast DNS and DNS-SD for
Linux, and likely for some BSD systems as well. This technology is
essential for automatic network configuration, service discovery on
local networks, and driverless printing and scanning. It also can be
useful for peer service discovery in a cloud environment.

# Package philosophy

The Avahi API wrapper, provided by this package, attempts to be as close
to the original Avahi C API and as transparent as possible. However,
the following differences still exist:
  - Events are reported via channels, not via callbacks as in C
  - AvahiPoll object is not exposed and handled internally
  - A workaround for Avahi localhost handling bug is provided (for details,
    see the "Loopback interface handling and localhost" section below)

# A bit of theory (Multicast DNS and DNS-SD essentials)

The Avahi API is much simpler to understand when the reader knows the
basics of the Multicast DNS and DNS-SD protocols.

DNS is a kind of distributed key-value database. In classical (unicast)
DNS, records are maintained by a hierarchy of servers, while in
Multicast DNS (mDNS) each participating host maintains its own records
and responds when queried, usually using multicast UDP as transport.

With classical DNS, clients perform database queries by contacting DNS
servers. In contrast, with mDNS, clients send their queries to all other
hosts on the local network using UDP multicast (e.g., "I need an IP address
for 'example.local'. Who knows the answer?"). The hosts then respond
individually. To improve efficiency, when a new host connects to the network,
it announces its resource records (RRs) to all interested parties and
attempts to notify others before it disconnects. Clients can capture and
cache this information, eliminating the need for a slow network query each
time this information is required.

Each entry in the DNS database (called a Resource Record, RR) is
identified by a search key consisting of:
  - record name
  - record class
  - record type

The record name always looks like a domain name, i.e., it is a string
consisting of dot-separated labels. The name "example.com" consists of
two labels: "example" and "com".

This syntax is used even for names that are not domains by themselves.
For example, "1.0.0.127.in-addr.arpa" is the IP address "127.0.0.1", written
using DNS name syntax (notice the reverse order of labels), and
"_http._tcp.local" is the collective name of all HTTP servers running over
TCP on the local network.

To distinguish between normal domains and pseudo-domains, a number of special
top-level domains have been reserved, such as "in-addr.arpa" for IP addresses.

DNS defines many classes, but the only class relevant to mDNS is IN,
which stands for "Internet". That's all there is to it.

Record type is more important, as many record types are used.
The most important for us are the following:

	A       - contains one or more IPv4 addresses
	AAAA    - contains one or more IPv6 addresses
	PTR     - pointer record, pointing to some other domain name
	SRV     - service descriptor
	TXT     - contains additional information as a list of key=value
	          textual pairs

Once we have a record name and type, we can query the record value.
Interpretation of this value depends on the record type.

Now let's manually discover all IPP printers on our local network.
We'll use the small utility [mcdig], which allows manual execution of
Multicast DNS queries.

First, let's list all services on the local network. This is a query
of the "_services._dns-sd._udp.local" records of type PTR, and [mcdig]
will return the following answer (shortened):

	$ mcdig _services._dns-sd._udp.local ptr
	;; ANSWER SECTION:
	_services._dns-sd._udp.local.   4500    IN      PTR     _http._tcp.local.
	_services._dns-sd._udp.local.   4500    IN      PTR     _https._tcp.local.
	_services._dns-sd._udp.local.   4500    IN      PTR     _ipp._tcp.local.
	_services._dns-sd._udp.local.   4500    IN      PTR     _ipps._tcp.local.
	_services._dns-sd._udp.local.   4500    IN      PTR     _printer._tcp.local.

This is the same list returned by `avahi-browse -a`, and programmatically
it can be obtained using the [ServiceTypeBrowser] object.

Note that "_services._dns-sd._udp.<domain>" is the reserved name for this
purpose, and <domain> is usually "local" (this top-level domain is reserved
for mDNS).

From the output, we see that someone on the network provides the
"_ipp._tcp.local." service (IPP printing), "_http._tcp.local." service
(HTTP server), and so on. In a typical network, there will be many
services that may appear multiple times in the answer.

Now, since we're only interested in IPP printers:

	$ mcdig _ipp._tcp.local. ptr
	;; ANSWER SECTION:
	_ipp._tcp.local.   4500    IN     PTR     Kyocera\ ECOSYS\ M2040dn._ipp._tcp.local.

Now we have a service instance name, "Kyocera ECOSYS M2040dn".
Unlike classical DNS, mDNS labels may contain spaces (and virtually any
valid UTF-8 characters). Although these labels look like human-readable
names, they are network-unique (enforced by the protocol) and can be used
to unambiguously identify the device.

The same list would be returned by `avahi-browse _ipp._tcp` (note that the
.local suffix is implied) or using the [ServiceBrowser] object.

Now we need to know more about the device, so the next two queries are:

	$ mcdig Kyocera\ ECOSYS\ M2040dn._ipp._tcp.local. srv
	Kyocera\ ECOSYS\ M2040dn._ipp._tcp.local.   120    IN    SRV    0 0 631 KM7B6A91.local.

	$ mcdig Kyocera\ ECOSYS\ M2040dn._ipp._tcp.local. txt
	Kyocera\ ECOSYS\ M2040dn._ipp._tcp.local.   4500   IN    TXT    "txtvers=1" "pdl=image/pwg-raster,..." ...

This provides the following information:

  - SRV record contains a hostname (which is not the same as the
    instance name, and often is not as friendly or human-readable)
    and IP port (631, the third parameter in the SRV RR)

  - TXT record contains many "key=value" pairs describing various
    characteristics of the device.

The final step is to obtain the device's IP addresses. For this, we need
the hostname obtained in the previous steps:

	$ mcdig KM7B6A91.local. a
	KM7B6A91.local.	120     IN      A       192.168.1.102

	$ mcdig KM7B6A91.local. aaaa
	KM7B6A91.local.	120     IN      AAAA    fe80::217:c8ff:fe7b:6a91

The response is significantly shortened here. The TXT record is omitted
entirely as it's quite large.

The complete discovery process looks like this:

	INPUT: "_ipp._tcp.local."                               (the service type)
	 |
	 --> Query PTR record
	      |
	      --> "Kyocera ECOSYS M2040dn._ipp._tcp.local."     (the instance name)
	          |
	          |-> Query SRV record
	          |    |
	          |    |-> 631                                  (TCP port)
	          |    |
	          |    --> "KM7B6A91.local."                    (the hostname)
	          |          |
	          |          |-> Query A record
	          |          |    |
	          |          |    --> 192.168.1.102             (IPv4 address)
	          |          |
	          |          --> Query AAAA record
	          |               |
	          |               --> fe80::217:c8ff:fe7b:6a91  (IPv6 address)
	          |
	          -> Query TXT record
	              |
	              --> A lot of key=value pairs              (device description)

This information can be obtained programmatically using the [ServiceResolver]
object, which performs all these steps under the hood.

Finally, we can look up IP addresses by hostname and hostname by IP address:

	$ mcdig KM7B6A91.local. a
	;; ANSWER SECTION:
	KM7B6A91.local.                120  IN    A       192.168.1.102

	$ mcdig 102.1.168.192.in-addr.arpa ptr
	;; ANSWER SECTION:
	102.1.168.192.in-addr.arpa.    120  IN    PTR     KM7B6A91.local.

This corresponds to the avahi commands `avahi-resolve-host-name KM7B6A91.local`
and `avahi-resolve-address 192.168.1.102`.

The [HostNameResolver] and [AddressResolver] objects provide similar
functionality through an API.

# Key objects

The key objects exposed by this package are:

  - [Client] represents a client connection to the avahi-daemon
  - Browsers: [DomainBrowser], [RecordBrowser], [ServiceBrowser],
    [ServiceTypeBrowser]
  - Resolvers: [AddressResolver], [HostNameResolver], [ServiceResolver]
  - [EntryGroup] implements the Avahi publishing API

These objects have a 1:1 relationship with the corresponding avahi objects
(i.e., Client represents AvahiClient, DomainBrowser represents AvahiDomainBrowser,
and so on).

Objects are explicitly created with constructor functions (e.g., [NewClient],
[NewDomainBrowser], [NewServiceResolver], etc.).

All these objects report their state changes and discovered information
through a provided channel (use the Chan() method to obtain the channel).
There is also a [context.Context]-aware Get() method that can be used to
wait for the next event.

Since these objects own resources (such as a DBus connection to the
avahi-daemon) that are not automatically released when objects are
garbage-collected, it's important to call the appropriate Close method
when an object is no longer in use.

Once an object is closed, the sending side of its event channel is closed
too, which unblocks all users waiting for events.

# Client

The [Client] represents a client connection to the avahi-daemon.
Client is a required parameter for creating Browsers and Resolvers, and
"owns" these objects.

Client has a state that can change dynamically. Changes in the Client
state are reported as a series of [ClientEvent] events through the
[Client.Chan] channel or the [Client.Get] convenience wrapper.

The Client manages the underlying AvahiPoll object (Avahi event loop)
automatically and doesn't expose it through its interface.

# Browsers

A Browser constantly monitors the network for newly discovered or removed
objects of a specified type and reports discovered information as a series
of events delivered through a provided channel.

More technically, a browser monitors the network for reception of
mDNS messages of a browser-specific type and reports these messages
as browser events.

There are 5 types of browser events, represented as values of the
[BrowserEvent] type:
  - [BrowserNew] - a new object was discovered on the network
  - [BrowserRemove] - an object was removed from the network
  - [BrowserCacheExhausted] - a one-time hint event notifying the user
    that all entries from the avahi-daemon cache have been sent
  - [BrowserAllForNow] - a one-time hint event notifying the user that
    more events are unlikely to appear in the near future
  - [BrowserFailure] - browsing failed and needs to be restarted

Avahi documentation doesn't explain in detail when [BrowserAllForNow]
is generated, but generally, it's generated after a one-second interval
from the reception of the last mDNS message of the related type has elapsed.

Each browser has a constructor function (e.g., [NewDomainBrowser]) and
three methods:
  - Chan() returns the event channel
  - Get() is a convenience wrapper that waits for the next event
    and can be canceled using a [context.Context] parameter
  - Close() closes the browser

It's important to call Close() when a browser is no longer in use.

# Resolvers

A Resolver performs a series of appropriate mDNS queries to resolve
supplied parameters into the requested information, depending on the
Resolver type (e.g., ServiceResolver resolves a service name into a
hostname, IP address:port, and TXT record).

Like Browsers, Resolvers return discovered information as a series of
resolver events.

There are 2 types of resolver events, represented by integer values
of the [ResolverEvent] type:
  - [ResolverFound] - a new portion of required information has been
    received from the network
  - [ResolverFailure] - resolving failed and needs to be restarted

Note that a single query may return multiple [ResolverFound] events.
For example, if the target has multiple IP addresses, each address will
be reported in a separate event.

Unlike the Browser, the Resolver does not provide any indication of
which event is considered "last" in the sequence. Technically, there is
no definitive "last" event, as a continuously running Resolver will
generate a [ResolverFound] event each time the service data changes.
However, if we simply need to connect to a discovered service, we must
eventually stop waiting. A reasonable approach would be to wait for a
meaningful duration (e.g., 1 second) after the last event in the
sequence arrives.

# EntryGroup

[EntryGroup] implements the Avahi publishing API. It is essentially
a collection of resource entries that can be published "atomically",
i.e., either the entire group is published or none of it is.

Records can be added to the EntryGroup using [EntryGroup.AddService],
[EntryGroup.AddAddress], and [EntryGroup.AddRecord] methods. Existing
services can be modified using [EntryGroup.AddServiceSubtype] and
[EntryGroup.UpdateServiceTxt]. Once the group is configured, the
application must call [EntryGroup.Commit] for changes to take effect.

When records are added (even before Commit), Avahi performs some basic
consistency checking of the group. If consistency is violated or
added records contain invalid data, the appropriate call will fail
with a suitable error code.

When publishing services, there is no way to set the service IP address
explicitly. Instead, Avahi deduces the appropriate IP address based on
the network interface being used and the available addresses assigned
to that interface.

Like other objects, EntryGroup maintains a dynamic state and reports
its state changes using [EntryGroupEvent], which can be received either
through the channel returned by [EntryGroup.Chan] or via the
[EntryGroup.Get] convenience wrapper.

As required by the protocol, EntryGroup performs conflict checking,
which takes some time. As a result of this process, the EntryGroup will
eventually reach either [EntryGroupStateEstablished] or
[EntryGroupStateCollision] state.

Unfortunately, in the case of a collision, there is no detailed reporting
about which entry caused the collision. Therefore, it's not recommended
to mix unrelated entries in the same group.

# IPv4 vs IPv6

When creating a new Browser or Resolver, the protocol parameter of the
constructor function specifies the transport protocol used for queries.

Some Resolver constructors have a second parameter of the [Protocol]
type, called "addrproto". This parameter specifies which kind of
addresses (IPv4 or IPv6) we are interested in receiving (technically,
which kind of address records, A or AAAA, are queried).

If you create a Browser with [ProtocolUnspec] as the transport protocol,
it will report both IPv4 and IPv6 RRs as separate events.

A new Resolver created with [ProtocolUnspec] as the transport protocol will
use IPv6 as its transport protocol, as if [ProtocolIP6] were specified.

If "addrproto" is specified as [ProtocolUnspec], the Resolver will always
query for addresses that match the transport protocol.

This behavior can be summarized in the following table:

	proto           addrproto       transport       query for
	----------------------------------------------------------
	ProtocolIP4     ProtocolIP4     IPv4            IPv4
	ProtocolIP4     ProtocolIP6     IPv4            IPv6
	ProtocolIP4     ProtocolUnspec  IPv4            IPv4

	ProtocolIP6     ProtocolIP4     IPv6            IPv4
	ProtocolIP6     ProtocolIP6     IPv6            IPv6
	ProtocolIP6     ProtocolUnspec  IPv6            IPv6

	ProtocolUnspec  ProtocolIP4     IPv6            IPv4
	ProtocolUnspec  ProtocolIP6     IPv6            IPv6
	ProtocolUnspec  ProtocolUnspec  IPv6            IPv6

By default, the Avahi daemon publishes both IPv4 and IPv6 addresses when
queried over IPv4, but only IPv6 addresses when queried over IPv6. This
default can be changed using the 'publish-aaaa-on-ipv4' and
'publish-a-on-ipv6' options in 'avahi-daemon.conf'.

Other servers (especially DNS-SD servers found on devices like printers
or scanners) may have different, sometimes surprising, behavior.

Therefore, it makes sense to perform queries for all four
transport/address combinations and merge the results.

# Loopback interface handling and localhost

Since the loopback network interface doesn't support multicasting,
Avahi simply emulates the appropriate functionality.

Loopback support is essential for implementing the [IPP over USB]
protocol, and the [ipp-usb] daemon actively uses it. It allows many
modern printers and scanners to work seamlessly under Linux.

Unfortunately, loopback support is broken in Avahi. This is a long
story, but in short:
  - Services published at loopback addresses (127.0.0.1 or ::1)
    are erroneously reported by AvahiServiceResolver as being
    published at the real hostname and domain, instead of
    "localhost.localdomain"
  - AvahiAddressResolver also resolves these addresses into the
    real hostname and domain instead of "localhost.localdomain"
  - AvahiHostNameResolver fails to resolve either "localhost" or
    "localhost.localdomain" host names

This library provides a workaround, but it needs to be explicitly
enabled using the [ClientLoopbackWorkarounds] flag:

	clnt, err := NewClient(ClientLoopbackWorkarounds)

If this flag is used, the following changes occur:
  - [ServiceResolver] and [AddressResolver] will return "localhost.localdomain"
    for loopback addresses
  - [HostNameResolver] will resolve "localhost" and "localhost.localdomain"
    as either 127.0.0.1 or ::1, depending on the value of the
    proto parameter in the [NewHostNameResolver] call. Note that if
    proto is [ProtocolUnspec], NewHostNameResolver will use
    [ProtocolIP6] by default to be consistent with other Avahi API
    (see the "IPv4 vs IPv6" section for details).

# IP addresses

This package uniformly uses [netip.Addr] to represent addresses. Unlike
[net.Addr], this format is compact, convenient, and comparable.

When addresses are received from Avahi (for example, as part of a
[ServiceResolverEvent]), the following rules apply:
  - IPv4 addresses are represented as 4-byte netip.Addr, not as
    16-byte IPv6-mapped IPv4 addresses
  - Link-local IPv6 addresses include a zone in symbolic format
    (e.g., fe80::1ff:fe23:4567:890a%eth2, not fe80::1ff:fe23:4567:890a%3).
    If the symbolic zone name cannot be obtained, a fallback to the
    numeric format will occur. This behavior is consistent with the
    Go standard library.

When an address is sent from the application to Avahi, the following
rules apply:
  - Both genuine IPv4 and IPv6-mapped IPv4 addresses are accepted
  - For IPv6 addresses, the zone is ignored

# The Poller

[Poller] is a helper object that simplifies the event loop when working
with many instances of Browsers and Resolvers (the typical case for
Avahi programming).

Poller is not part of the native Avahi API and is added here for convenience.

Poller allows you to "connect" many event sources to a single object
and use its [Poller.Poll] method to gather events from all connected
objects.

See the project's README.md for a usage example.

[IPP over USB]: https://www.usb.org/document-library/ipp-protocol-10
[ipp-usb]: https://github.com/OpenPrinting/ipp-usb
[mcdig]: https://github.com/alexpevzner/mcdig

[Avahi]: https://avahi.org/
*/
package avahi
