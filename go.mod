module github.com/go-i2p/go-sam-go

go 1.26.3

require (
	github.com/go-i2p/common v0.1.60000-0.20260701134558-e5f5cf65a7f5
	github.com/go-i2p/crypto v0.1.60000-0.20260701135847-3ade996b68a0
	github.com/go-i2p/i2pkeys v0.33.92
	github.com/go-i2p/logger v0.1.60000-0.20260701134448-2648c3b0e040
	github.com/samber/oops v1.23.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/oklog/ulid/v2 v2.1.2 // indirect
	github.com/samber/lo v1.53.0 // indirect
	github.com/sirupsen/logrus v1.10.2 // indirect
	go.opentelemetry.io/otel v1.46.0 // indirect
	go.opentelemetry.io/otel/trace v1.46.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
)

retract (
	v0.1.59999
	v0.1.5999
	v0.1.599
)
