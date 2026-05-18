//go:build darwin

package nwtransport

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/pion/transport/v4"
	"github.com/tmc/apple/x/network/nwpacket"
)

func TestNetImplementsTransportNet(t *testing.T) {
	var _ transport.Net = (*Net)(nil)
}

func TestNativePacketAddress(t *testing.T) {
	tests := []struct {
		name    string
		network string
		address string
		want    bool
	}{
		{"udp4 host", "udp4", "192.0.2.1:0", true},
		{"udp6 scoped host", "udp6", "[fe80::1%awdl0]:9", true},
		{"unspecified", "udp4", "0.0.0.0:0", false},
		{"multicast", "udp4", "224.0.0.251:5353", false},
		{"tcp", "tcp4", "192.0.2.1:0", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got, err := nativePacketAddress(tt.network, tt.address)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("nativePacketAddress(%q, %q) native = %t, want %t", tt.network, tt.address, got, tt.want)
			}
		})
	}
}

func TestListenPacketUsesConfiguredAddressForWildcard(t *testing.T) {
	n, err := New(Config{Packet: nwpacket.Config{
		LocalAddr: &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 1000},
	}})
	if err != nil {
		t.Fatal(err)
	}
	var gotLocal *net.UDPAddr
	n.listenPacket = func(ctx context.Context, config nwpacket.Config) (net.PacketConn, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		gotLocal = copyUDPAddr(config.LocalAddr)
		return &queuePacketConn{}, nil
	}
	if _, err := n.ListenPacket("udp4", "0.0.0.0:9999"); err != nil {
		t.Fatal(err)
	}
	if gotLocal == nil || gotLocal.String() != "192.0.2.1:9999" {
		t.Fatalf("native local = %v, want 192.0.2.1:9999", gotLocal)
	}
}

func TestListenUDPUsesConfiguredAddressForNilLocal(t *testing.T) {
	n, err := New(Config{Packet: nwpacket.Config{
		InterfaceName: "awdl0",
		LocalAddr:     &net.UDPAddr{IP: net.ParseIP("fe80::1"), Port: 1000},
	}})
	if err != nil {
		t.Fatal(err)
	}
	var gotLocal *net.UDPAddr
	n.listenPacket = func(ctx context.Context, config nwpacket.Config) (net.PacketConn, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		gotLocal = copyUDPAddr(config.LocalAddr)
		return &queuePacketConn{}, nil
	}
	if _, err := n.ListenUDP("udp6", nil); err != nil {
		t.Fatal(err)
	}
	if gotLocal == nil || gotLocal.String() != "[fe80::1%awdl0]:1000" {
		t.Fatalf("native local = %v, want [fe80::1%%awdl0]:1000", gotLocal)
	}
}

func TestUnzoneLinkLocalAddr(t *testing.T) {
	linkLocal := net.ParseIP("fe80::1")
	addr := unzoneLinkLocalAddr(&net.IPNet{IP: linkLocal, Mask: net.CIDRMask(64, 128)})
	ipAddr, ok := addr.(*net.IPAddr)
	if !ok {
		t.Fatalf("unzoneLinkLocalAddr returned %T, want *net.IPAddr", addr)
	}
	if !ipAddr.IP.Equal(linkLocal) {
		t.Fatalf("IP = %s, want %s", ipAddr.IP, linkLocal)
	}
	if ipAddr.Zone != "" {
		t.Fatalf("Zone = %q, want empty", ipAddr.Zone)
	}

	ipv4 := &net.IPNet{IP: net.ParseIP("192.0.2.1"), Mask: net.CIDRMask(24, 32)}
	if got := unzoneLinkLocalAddr(ipv4); got != ipv4 {
		t.Fatalf("IPv4 address was changed")
	}
}

func TestNativeDialUDPAddrs(t *testing.T) {
	n := &Net{config: Config{Packet: nwpacket.Config{
		InterfaceName: "awdl0",
		LocalAddr:     &net.UDPAddr{IP: net.ParseIP("fe80::1"), Port: 1000},
	}}}
	local, remote, ok := n.nativeDialUDPAddrs("udp6", nil, &net.UDPAddr{IP: net.ParseIP("fe80::2"), Port: 2000})
	if !ok {
		t.Fatal("nativeDialUDPAddrs rejected link-local dial")
	}
	if local.String() != "[fe80::1%awdl0]:1000" {
		t.Fatalf("local = %s, want [fe80::1%%awdl0]:1000", local)
	}
	if remote.String() != "[fe80::2%awdl0]:2000" {
		t.Fatalf("remote = %s, want [fe80::2%%awdl0]:2000", remote)
	}
	if n.config.Packet.LocalAddr.Zone != "" {
		t.Fatalf("config local addr was mutated: %s", n.config.Packet.LocalAddr)
	}
}

func TestNativeDialUDPAddrsRejectsFallbackCases(t *testing.T) {
	n := &Net{config: Config{Packet: nwpacket.Config{
		LocalAddr: &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 1000},
	}}}
	tests := []struct {
		name    string
		network string
		laddr   *net.UDPAddr
		raddr   *net.UDPAddr
	}{
		{
			name:    "tcp",
			network: "tcp",
			raddr:   &net.UDPAddr{IP: net.ParseIP("192.0.2.2"), Port: 2000},
		},
		{
			name:    "missing local",
			network: "udp4",
			laddr:   nil,
			raddr:   &net.UDPAddr{IP: net.ParseIP("192.0.2.2"), Port: 2000},
		},
		{
			name:    "unspecified remote",
			network: "udp4",
			laddr:   &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 1000},
			raddr:   &net.UDPAddr{IP: net.IPv4zero, Port: 2000},
		},
		{
			name:    "family mismatch",
			network: "udp4",
			laddr:   &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 1000},
			raddr:   &net.UDPAddr{IP: net.ParseIP("fe80::2"), Port: 2000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testNet := n
			if tt.name == "missing local" {
				testNet = &Net{}
			}
			if _, _, ok := testNet.nativeDialUDPAddrs(tt.network, tt.laddr, tt.raddr); ok {
				t.Fatal("nativeDialUDPAddrs accepted fallback case")
			}
		})
	}
}

func TestFallbackHandlesNonNativePolicy(t *testing.T) {
	fallback := &fallbackNet{}
	n, err := New(Config{Fallback: fallback})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := n.ListenPacket("udp4", "0.0.0.0:0"); !errors.Is(err, errFallback) {
		t.Fatalf("ListenPacket err = %v, want %v", err, errFallback)
	}
	if _, err := n.ListenUDP("udp4", nil); !errors.Is(err, errFallback) {
		t.Fatalf("ListenUDP err = %v, want %v", err, errFallback)
	}
	if _, err := n.DialUDP("udp4", nil, &net.UDPAddr{IP: net.ParseIP("192.0.2.2"), Port: 9}); !errors.Is(err, errFallback) {
		t.Fatalf("DialUDP err = %v, want %v", err, errFallback)
	}
	if _, err := n.Dial("tcp", "example.com:443"); !errors.Is(err, errFallback) {
		t.Fatalf("Dial err = %v, want %v", err, errFallback)
	}
	if _, err := n.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4zero, Port: 0}); !errors.Is(err, errFallback) {
		t.Fatalf("ListenTCP err = %v, want %v", err, errFallback)
	}
	if _, err := n.DialTCP("tcp4", nil, &net.TCPAddr{IP: net.ParseIP("192.0.2.2"), Port: 443}); !errors.Is(err, errFallback) {
		t.Fatalf("DialTCP err = %v, want %v", err, errFallback)
	}
	if _, err := n.ResolveUDPAddr("udp", "stun.example.test:3478"); !errors.Is(err, errFallback) {
		t.Fatalf("ResolveUDPAddr err = %v, want %v", err, errFallback)
	}
	if _, err := n.CreateDialer(&net.Dialer{}).Dial("tcp", "example.com:443"); !errors.Is(err, errFallback) {
		t.Fatalf("CreateDialer Dial err = %v, want %v", err, errFallback)
	}
	if _, err := n.CreateListenConfig(&net.ListenConfig{}).Listen(context.Background(), "tcp4", "0.0.0.0:0"); !errors.Is(err, errFallback) {
		t.Fatalf("CreateListenConfig Listen err = %v, want %v", err, errFallback)
	}
	if _, err := n.CreateListenConfig(&net.ListenConfig{}).ListenPacket(context.Background(), "udp4", "0.0.0.0:0"); !errors.Is(err, errFallback) {
		t.Fatalf("CreateListenConfig ListenPacket err = %v, want %v", err, errFallback)
	}

	want := []string{
		"ListenPacket",
		"ListenUDP",
		"DialUDP",
		"Dial",
		"ListenTCP",
		"DialTCP",
		"ResolveUDPAddr",
		"CreateDialer",
		"Dialer.Dial",
		"CreateListenConfig",
		"ListenConfig.Listen",
		"CreateListenConfig",
		"ListenConfig.ListenPacket",
	}
	if !equalStrings(fallback.calls, want) {
		t.Fatalf("fallback calls = %q, want %q", fallback.calls, want)
	}
}

func TestCreateDialerUsesNativeUDP(t *testing.T) {
	n, err := New(Config{Packet: nwpacket.Config{
		LocalAddr: &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 1000},
	}})
	if err != nil {
		t.Fatal(err)
	}
	var gotLocal *net.UDPAddr
	n.listenPacket = func(ctx context.Context, config nwpacket.Config) (net.PacketConn, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		gotLocal = copyUDPAddr(config.LocalAddr)
		return &queuePacketConn{}, nil
	}
	conn, err := n.CreateDialer(&net.Dialer{}).Dial("udp4", "192.0.2.2:2000")
	if err != nil {
		t.Fatal(err)
	}
	if gotLocal == nil || gotLocal.String() != "192.0.2.1:1000" {
		t.Fatalf("native local = %v, want 192.0.2.1:1000", gotLocal)
	}
	if conn.RemoteAddr().String() != "192.0.2.2:2000" {
		t.Fatalf("remote = %s, want 192.0.2.2:2000", conn.RemoteAddr())
	}
}

func TestListenConfigListenPacketCanceledContext(t *testing.T) {
	n, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = n.CreateListenConfig(&net.ListenConfig{}).ListenPacket(ctx, "udp4", "127.0.0.1:0")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ListenPacket canceled context err = %v, want %v", err, context.Canceled)
	}
}

func TestUDPConnConnectedReadWrite(t *testing.T) {
	remote := &net.UDPAddr{IP: net.ParseIP("192.0.2.2"), Port: 2000}
	other := &net.UDPAddr{IP: net.ParseIP("192.0.2.3"), Port: 2000}
	packets := []queuedPacket{
		{data: []byte("skip"), addr: other},
		{data: []byte("ok"), addr: remote},
	}
	packetConn := &queuePacketConn{packets: packets}
	conn := &udpConn{PacketConn: packetConn, remote: remote}

	buf := make([]byte, 16)
	n, addr, err := conn.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(buf[:n]); got != "ok" {
		t.Fatalf("read payload = %q, want ok", got)
	}
	if !sameUDPAddr(addr.(*net.UDPAddr), remote) {
		t.Fatalf("read addr = %s, want %s", addr, remote)
	}

	if _, err := conn.WriteTo([]byte("bad"), other); !errors.Is(err, transport.ErrNoAddressAssigned) {
		t.Fatalf("WriteTo other err = %v, want %v", err, transport.ErrNoAddressAssigned)
	}
	if n, err := conn.Write([]byte("ping")); err != nil || n != 4 {
		t.Fatalf("Write = %d, %v; want 4, nil", n, err)
	}
	if len(packetConn.writes) != 1 || !sameUDPAddr(packetConn.writes[0].addr, remote) {
		t.Fatalf("writes = %#v, want one write to %s", packetConn.writes, remote)
	}
}

type queuedPacket struct {
	data []byte
	addr *net.UDPAddr
}

type queuePacketConn struct {
	packets []queuedPacket
	writes  []queuedPacket
}

func (c *queuePacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	if len(c.packets) == 0 {
		return 0, nil, errors.New("no packet")
	}
	pkt := c.packets[0]
	c.packets = c.packets[1:]
	return copy(b, pkt.data), pkt.addr, nil
}

func (c *queuePacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	udpAddr, ok := addr.(*net.UDPAddr)
	if !ok {
		return 0, transport.ErrNotUDPAddress
	}
	c.writes = append(c.writes, queuedPacket{
		data: append([]byte(nil), b...),
		addr: copyUDPAddr(udpAddr),
	})
	return len(b), nil
}

func (c *queuePacketConn) Close() error                     { return nil }
func (c *queuePacketConn) LocalAddr() net.Addr              { return &net.UDPAddr{} }
func (c *queuePacketConn) SetDeadline(time.Time) error      { return nil }
func (c *queuePacketConn) SetReadDeadline(time.Time) error  { return nil }
func (c *queuePacketConn) SetWriteDeadline(time.Time) error { return nil }

var errFallback = errors.New("fallback called")

type fallbackNet struct {
	calls []string
}

func (n *fallbackNet) record(call string) {
	n.calls = append(n.calls, call)
}

func (n *fallbackNet) ListenPacket(string, string) (net.PacketConn, error) {
	n.record("ListenPacket")
	return nil, errFallback
}

func (n *fallbackNet) ListenUDP(string, *net.UDPAddr) (transport.UDPConn, error) {
	n.record("ListenUDP")
	return nil, errFallback
}

func (n *fallbackNet) ListenTCP(string, *net.TCPAddr) (transport.TCPListener, error) {
	n.record("ListenTCP")
	return nil, errFallback
}

func (n *fallbackNet) Dial(string, string) (net.Conn, error) {
	n.record("Dial")
	return nil, errFallback
}

func (n *fallbackNet) DialUDP(string, *net.UDPAddr, *net.UDPAddr) (transport.UDPConn, error) {
	n.record("DialUDP")
	return nil, errFallback
}

func (n *fallbackNet) DialTCP(string, *net.TCPAddr, *net.TCPAddr) (transport.TCPConn, error) {
	n.record("DialTCP")
	return nil, errFallback
}

func (n *fallbackNet) ResolveIPAddr(string, string) (*net.IPAddr, error) {
	n.record("ResolveIPAddr")
	return nil, errFallback
}

func (n *fallbackNet) ResolveUDPAddr(string, string) (*net.UDPAddr, error) {
	n.record("ResolveUDPAddr")
	return nil, errFallback
}

func (n *fallbackNet) ResolveTCPAddr(string, string) (*net.TCPAddr, error) {
	n.record("ResolveTCPAddr")
	return nil, errFallback
}

func (n *fallbackNet) Interfaces() ([]*transport.Interface, error) {
	n.record("Interfaces")
	return nil, errFallback
}

func (n *fallbackNet) InterfaceByIndex(int) (*transport.Interface, error) {
	n.record("InterfaceByIndex")
	return nil, errFallback
}

func (n *fallbackNet) InterfaceByName(string) (*transport.Interface, error) {
	n.record("InterfaceByName")
	return nil, errFallback
}

func (n *fallbackNet) CreateDialer(*net.Dialer) transport.Dialer {
	n.record("CreateDialer")
	return fallbackDialer{net: n}
}

func (n *fallbackNet) CreateListenConfig(*net.ListenConfig) transport.ListenConfig {
	n.record("CreateListenConfig")
	return fallbackListenConfig{net: n}
}

type fallbackDialer struct {
	net *fallbackNet
}

func (d fallbackDialer) Dial(string, string) (net.Conn, error) {
	d.net.record("Dialer.Dial")
	return nil, errFallback
}

type fallbackListenConfig struct {
	net *fallbackNet
}

func (c fallbackListenConfig) Listen(context.Context, string, string) (net.Listener, error) {
	c.net.record("ListenConfig.Listen")
	return nil, errFallback
}

func (c fallbackListenConfig) ListenPacket(context.Context, string, string) (net.PacketConn, error) {
	c.net.record("ListenConfig.ListenPacket")
	return nil, errFallback
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
