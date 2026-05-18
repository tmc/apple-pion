# apple-pion

This module contains small Pion adapters for `github.com/tmc/apple`.

The current packages are:

- `github.com/tmc/apple-pion/icepolicy`: helpers for explicit link-local ICE
  candidate publication when trusted peers cannot rely on mDNS candidate
  exchange.
- `github.com/tmc/apple-pion/nwtransport`: a Pion `transport.Net` adapter that
  routes concrete UDP listeners, configured wildcard UDP listeners, and UDP
  dials through `github.com/tmc/apple/x/network/nwpacket`, while falling back to
  Pion's standard network implementation for DNS, TCP, unconstrained wildcard
  UDP, and unsupported families.

The module uses released `github.com/tmc/apple v0.6.6`, which includes
`x/network/nwpacket`.

`nwtransport` is intentionally hybrid. Network.framework owns the UDP listeners
and UDP dials needed for Pion ICE on a selected Apple link. DNS, TCP,
unconstrained wildcard UDP, TURN/STUN helper traffic outside that selected UDP
surface, and unsupported address families stay on the fallback `transport.Net`.
