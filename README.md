# apple-pion

This module contains small Pion adapters for `github.com/tmc/apple`.

The current package is:

- `github.com/tmc/apple-pion/nwtransport`: a Pion `transport.Net` adapter that
  routes concrete UDP listeners and UDP dials through
  `github.com/tmc/apple/network/nwpacket`, while falling back to Pion's standard
  network implementation for DNS, TCP, wildcard UDP, and unsupported families.

The module uses `replace github.com/tmc/apple => ../apple` until the
`network/nwpacket` package is available in a released `github.com/tmc/apple`
version.
