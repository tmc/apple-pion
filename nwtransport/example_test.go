//go:build darwin

package nwtransport_test

import (
	"fmt"
	"net"

	"github.com/pion/ice/v4"
	"github.com/pion/webrtc/v4"
	"github.com/tmc/apple-pion/icepolicy"
	"github.com/tmc/apple-pion/nwtransport"
	"github.com/tmc/apple/x/network/nwpacket"
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

func ExampleNew_pionSettingEngine() {
	localIP := net.ParseIP("fe80::1")
	netTransport, err := nwtransport.New(nwtransport.Config{
		Packet: nwpacket.Config{
			InterfaceName:     "awdl0",
			LocalAddr:         &net.UDPAddr{IP: localIP},
			IncludePeerToPeer: true,
			RequireInterface:  true,
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	var se webrtc.SettingEngine
	se.SetNet(netTransport)
	se.SetICEMulticastDNSMode(ice.MulticastDNSModeDisabled)
	icepolicy.Policy{RawHostCandidates: true}.Configure(&se, ice.MulticastDNSModeDisabled, localIP)

	api := webrtc.NewAPI(webrtc.WithSettingEngine(se))
	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer pc.Close()

	fmt.Println(true)

	// Output:
	// true
}
