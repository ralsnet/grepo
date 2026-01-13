module github.com/ralsnet/grepo/example

go 1.25.3

require (
	github.com/google/uuid v1.6.0
	github.com/ralsnet/grepo v0.0.0-00010101000000-000000000000
	github.com/ralsnet/grepo/cli v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
)

replace github.com/ralsnet/grepo => ../

replace github.com/ralsnet/grepo/cli => ../cli
