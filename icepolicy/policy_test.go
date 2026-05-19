package icepolicy

import (
	"net"
	"strings"
	"testing"

	"github.com/pion/webrtc/v4"
)

func TestPublishCandidateRawHostCandidates(t *testing.T) {
	policy := Policy{RawHostCandidates: true}
	candidate := "candidate:1 1 udp 2130706431 fd00::1 12345 typ host ufrag test"
	got := policy.PublishCandidate(candidate, net.ParseIP("fe80::1"))
	if !strings.Contains(got, " fe80::1 12345 typ host ") {
		t.Fatalf("host candidate was not rewritten: %s", got)
	}
}

func TestPublishCandidateIgnoresNonSyntheticCandidates(t *testing.T) {
	policy := Policy{RawHostCandidates: true}
	candidate := "candidate:1 1 udp 2130706431 10.0.0.1 12345 typ host ufrag test"
	if got := policy.PublishCandidate(candidate, net.ParseIP("10.0.0.1")); got != candidate {
		t.Fatalf("candidate changed unexpectedly: %s", got)
	}
}

func TestCandidateInitsFromSDP(t *testing.T) {
	policy := Policy{RawHostCandidates: true}
	sdp := strings.Join([]string{
		"v=0",
		"m=application 9 UDP/DTLS/SCTP webrtc-datachannel",
		"a=mid:0",
		"a=candidate:1 1 udp 2130706431 fd00::1 12345 typ host ufrag test",
		"a=end-of-candidates",
	}, "\n")
	candidates := CandidateInitsFromSDP(sdp, policy, net.ParseIP("fe80::1"))
	if len(candidates) != 1 {
		t.Fatalf("candidates = %d, want 1", len(candidates))
	}
	if !strings.Contains(candidates[0].Candidate, " fe80::1 12345 typ host ") {
		t.Fatalf("candidate was not rewritten: %s", candidates[0].Candidate)
	}
	if candidates[0].SDPMid == nil || *candidates[0].SDPMid != "0" {
		t.Fatalf("candidate mid = %v, want 0", candidates[0].SDPMid)
	}
	if candidates[0].SDPMLineIndex == nil || *candidates[0].SDPMLineIndex != 0 {
		t.Fatalf("candidate m-line = %v, want 0", candidates[0].SDPMLineIndex)
	}
}

func TestHostAddressRewriteRule(t *testing.T) {
	policy := Policy{RawHostCandidates: true}
	rule := policy.hostAddressRewriteRule(net.ParseIP("fe80::1"))
	if len(rule.External) != 1 || rule.External[0] != "fe80::1" {
		t.Fatalf("external = %v, want [fe80::1]", rule.External)
	}
	if rule.Local != "fe80::1" {
		t.Fatalf("local = %q, want fe80::1", rule.Local)
	}
	if rule.AsCandidateType != webrtc.ICECandidateTypeHost {
		t.Fatalf("candidate type = %s, want host", rule.AsCandidateType)
	}
	if rule.Mode != webrtc.ICEAddressRewriteReplace {
		t.Fatalf("mode = %v, want replace", rule.Mode)
	}
}
