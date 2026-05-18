package icepolicy_test

import (
	"fmt"
	"net"
	"strings"

	"github.com/pion/ice/v4"
	"github.com/pion/webrtc/v4"
	"github.com/tmc/apple-pion/icepolicy"
)

func Example() {
	policy := icepolicy.Policy{RawHostCandidates: true}
	fmt.Println(policy.UsesSyntheticHostCandidate(net.ParseIP("fe80::1")))

	// Output:
	// true
}

func ExamplePolicy_Configure() {
	var se webrtc.SettingEngine
	policy := icepolicy.Policy{RawHostCandidates: true}

	policy.Configure(&se, ice.MulticastDNSModeDisabled, net.ParseIP("fe80::1"))

	fmt.Println(policy.UsesSyntheticHostCandidate(net.ParseIP("fe80::1")))

	// Output:
	// true
}

func ExamplePolicy_PublishCandidate() {
	policy := icepolicy.Policy{RawHostCandidates: true}
	candidate := "candidate:1 1 udp 2130706431 fd00::1 12345 typ host"

	fmt.Println(policy.PublishCandidate(candidate, net.ParseIP("fe80::1")))

	// Output:
	// candidate:1 1 udp 2130706431 fe80::1 12345 typ host
}

func ExampleCandidateInitsFromSDP() {
	policy := icepolicy.Policy{RawHostCandidates: true}
	sdp := strings.Join([]string{
		"m=application 9 UDP/DTLS/SCTP webrtc-datachannel",
		"a=mid:0",
		"a=candidate:1 1 udp 2130706431 fd00::1 12345 typ host",
	}, "\n")

	candidates := icepolicy.CandidateInitsFromSDP(sdp, policy, net.ParseIP("fe80::1"))
	fmt.Println(candidates[0].Candidate)

	// Output:
	// candidate:1 1 udp 2130706431 fe80::1 12345 typ host
}

func ExampleStripSDPCandidates() {
	sdp := strings.Join([]string{
		"a=mid:0",
		"a=candidate:1 1 udp 2130706431 fd00::1 12345 typ host",
		"a=end-of-candidates",
	}, "\n")

	fmt.Println(icepolicy.StripSDPCandidates(sdp))

	// Output:
	// a=mid:0
}
