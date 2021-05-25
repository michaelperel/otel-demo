module github.com/michaelperel/otel-demo

go 1.16

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-redis/redis/v8 v8.8.3
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.20.0
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/exporters/otlp v0.20.0
	go.opentelemetry.io/otel/sdk v0.20.0
	go.opentelemetry.io/otel/trace v0.20.0
	google.golang.org/genproto v0.0.0-20200527145253-8367513e4ece // indirect
	google.golang.org/grpc v1.37.0
)
