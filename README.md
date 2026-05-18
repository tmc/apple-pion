# apple-pion

This module contains small Pion adapters for `github.com/tmc/apple`.

The current package is:

- `github.com/tmc/apple-pion/nwtransport`: a Pion `transport.Net` adapter that
  routes concrete UDP listeners, configured wildcard UDP listeners, and UDP
  dials through `github.com/tmc/apple/x/network/nwpacket`, while falling back to
  Pion's standard network implementation for DNS, TCP, unconstrained wildcard
  UDP, and unsupported families.

The module uses `replace github.com/tmc/apple => ../apple` until the
`x/network/nwpacket` package is available in a released `github.com/tmc/apple`
version.

`nwtransport` is intentionally hybrid. Network.framework owns the UDP listeners
and UDP dials needed for Pion ICE on a selected Apple link. DNS, TCP,
unconstrained wildcard UDP, TURN/STUN helper traffic outside that selected UDP
surface, and unsupported address families stay on the fallback `transport.Net`.
