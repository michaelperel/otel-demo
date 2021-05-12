package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/michaelperel/otel-demo/pkg/tracer_provider"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

var tr = otel.Tracer("cmd/server")

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate work to make viewing the span easier in the telemetry backend.
	time.Sleep(1 * time.Second)

	// Trace context propagated from the client request in r.Context().
	ctx := r.Context()
	_, span := tr.Start(
		ctx,
		"handle request",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	// Baggage from client propagates as well
	uk := attribute.Key("username")
	username := baggage.Value(ctx, uk)
	span.AddEvent("reading username baggage...", trace.WithAttributes(uk.String(username.AsString())))

	// Add the response as an attribute to the span. This can show up
	// differently in different backends, but in Jaeger attributes show up
	// as tags.
	response := "hello, world"
	span.SetAttributes(attribute.String("body", response))

	fmt.Fprint(w, response)
}

func main() {
	addr := "otel-agent:4317"

	shutdown := tracer_provider.Initialize(addr, "server")
	defer shutdown()

	otelHandler := otelhttp.NewHandler(http.HandlerFunc(helloHandler), "/hello")
	http.Handle("/hello", otelHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
