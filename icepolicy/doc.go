// Package icepolicy helps Pion publish explicit ICE host candidates for
// selected Apple link-local interfaces.
//
// Pion normally hides raw link-local host candidates behind mDNS. That is the
// right default, but Apple peer-to-peer links such as AWDL sometimes need
// explicit signaling between trusted peers. Policy configures Pion address
// rewrite rules to gather a synthetic host candidate and then publishes it as
// the selected link-local IP.
package icepolicy
