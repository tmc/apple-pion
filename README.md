# apple-pion

This module contains small Pion adapters for `github.com/tmc/apple`.

The current packages are:

- `github.com/tmc/apple-pion/icepolicy`: helpers for explicit link-local ICE
  candidate publication when trusted peers cannot rely on mDNS candidate
  exchange.
- `github.com/tmc/apple-pion/nwtransport`: a Pion `transport.Net` adapter that
  routes numeric UDP listeners, configured wildcard UDP listeners, and numeric
  UDP dials through `github.com/tmc/apple/x/network/nwpacket`, while falling
  back to Pion's standard network implementation for DNS, TCP, unconstrained
  wildcard UDP, and unsupported families.

The module uses released `github.com/tmc/apple v0.6.7`, which includes
`x/network/nwpacket` path reporting and outbound readiness retry knobs.

`nwtransport` is intentionally hybrid. Network.framework owns numeric UDP
listeners and numeric UDP dials needed for Pion ICE on a selected Apple link.
DNS, TCP, unconstrained wildcard UDP, TURN/STUN helper traffic outside that
selected UDP surface, and unsupported address families stay on the fallback
`transport.Net`.
