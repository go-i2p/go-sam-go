module github.com/go-i2p/go-sam-go

go 1.24.2

toolchain go1.24.4

require (
	github.com/go-i2p/common v0.0.0-20250819203334-e5459df35789
	github.com/go-i2p/crypto v0.0.0-20250822224541-85015740db11
	github.com/go-i2p/i2pkeys v0.33.92
	github.com/go-i2p/logger v0.0.0-20241123010126-3050657e5d0c
	github.com/samber/oops v1.19.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	github.com/samber/lo v1.51.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
)

replace github.com/go-i2p/go-sam-go => .
