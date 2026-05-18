//go:build darwin

package nwtransport_test

import (
	"fmt"
	"net"

	"github.com/tmc/apple-pion/nwtransport"
	"github.com/tmc/apple/network/nwpacket"
)

func ExampleNew() {
	netTransport, err := nwtransport.New(nwtransport.Config{
		Packet: nwpacket.Config{
			InterfaceName:     "awdl0",
			LocalAddr:         &net.UDPAddr{IP: net.ParseIP("fe80::1")},
			IncludePeerToPeer: true,
			RequireInterface:  true,
		},
	})
	fmt.Println(err == nil, netTransport != nil)

	// Output:
	// true true
}
