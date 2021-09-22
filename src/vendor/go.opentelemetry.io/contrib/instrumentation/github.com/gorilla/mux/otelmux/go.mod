module go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux

go 1.15

replace (
	go.opentelemetry.io/contrib => ../../../../..
	go.opentelemetry.io/contrib/propagators => ../../../../../propagators
)

require (
	github.com/felixge/httpsnoop v1.0.2
	github.com/gorilla/mux v1.8.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/contrib v0.22.0
	go.opentelemetry.io/contrib/propagators v0.22.0
	go.opentelemetry.io/otel v1.0.0-RC2
	go.opentelemetry.io/otel/oteltest v1.0.0-RC2
	go.opentelemetry.io/otel/trace v1.0.0-RC2
)
