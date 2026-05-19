package icepolicy

import (
	"net"
	"strings"

	"github.com/pion/ice/v4"
	"github.com/pion/webrtc/v4"
)

// Policy controls explicit host-candidate publication.
type Policy struct {
	RawHostCandidates bool
}

// Configure applies candidate publication settings to the Pion setting engine.
func (p Policy) Configure(se *webrtc.SettingEngine, mdnsMode ice.MulticastDNSMode, localIP net.IP) {
	if mdnsMode != ice.MulticastDNSModeDisabled || !p.UsesSyntheticHostCandidate(localIP) {
		return
	}
	rule := p.hostAddressRewriteRule(localIP)
	_ = se.SetICEAddressRewriteRules(rule)
}

// UsesSyntheticHostCandidate reports whether the policy needs Pion to gather a
// synthetic host candidate that will be published as localIP.
func (p Policy) UsesSyntheticHostCandidate(localIP net.IP) bool {
	return p.RawHostCandidates && localIP != nil && localIP.To4() == nil && localIP.IsLinkLocalUnicast()
}

func (p Policy) hostAddressRewriteRule(localIP net.IP) webrtc.ICEAddressRewriteRule {
	return webrtc.ICEAddressRewriteRule{
		External:        []string{syntheticHostCandidateIP(localIP)},
		Local:           localIP.String(),
		AsCandidateType: webrtc.ICECandidateTypeHost,
		Mode:            webrtc.ICEAddressRewriteReplace,
	}
}

// PublishCandidate rewrites one gathered host candidate to the selected local
// IP when the raw host-candidate policy is enabled.
func (p Policy) PublishCandidate(candidate string, localIP net.IP) string {
	if !p.UsesSyntheticHostCandidate(localIP) {
		return candidate
	}
	return rewriteHostCandidateAddressLine(candidate, localIP)
}

// CandidateInitsFromSDP returns explicit ICE candidates from sdp after applying
// policy to each host candidate. It does not modify the SDP.
func CandidateInitsFromSDP(sdp string, policy Policy, localIP net.IP) []webrtc.ICECandidateInit {
	var candidates []webrtc.ICECandidateInit
	var mid string
	var mline int
	for _, line := range strings.Split(sdp, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "m="):
			mline++
			mid = ""
		case strings.HasPrefix(line, "a=mid:"):
			mid = strings.TrimPrefix(line, "a=mid:")
		case strings.HasPrefix(line, "a=candidate:"):
			candidate := strings.TrimPrefix(line, "a=")
			init := webrtc.ICECandidateInit{Candidate: policy.PublishCandidate(candidate, localIP)}
			if mid != "" {
				midCopy := mid
				init.SDPMid = &midCopy
			}
			if mline > 0 {
				index := uint16(mline - 1)
				init.SDPMLineIndex = &index
			}
			candidates = append(candidates, init)
		}
	}
	return candidates
}

func syntheticHostCandidateIP(ip net.IP) string {
	if ip.To4() != nil {
		return "198.18.0.1"
	}
	return "fd00::1"
}

func rewriteHostCandidateAddressLine(line string, ip net.IP) string {
	if ip == nil {
		return line
	}
	prefix := ""
	candidate := line
	if strings.HasPrefix(candidate, "a=") {
		prefix = "a="
		candidate = strings.TrimPrefix(candidate, "a=")
	}
	if !strings.HasPrefix(candidate, "candidate:") {
		return line
	}
	fields := strings.Fields(candidate)
	if len(fields) < 8 || fields[7] != "host" {
		return line
	}
	fields[4] = ip.String()
	return prefix + strings.Join(fields, " ")
}
