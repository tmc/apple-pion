module github.com/tmc/apple-pion

go 1.25.0

replace github.com/tmc/apple => ../apple

require (
	github.com/pion/transport/v4 v4.0.1
	github.com/tmc/apple v0.6.3
)

require (
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/wlynxg/anet v0.0.5 // indirect
)
