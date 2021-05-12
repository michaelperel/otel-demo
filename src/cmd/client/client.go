package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/michaelperel/otel-demo/pkg/tracer_provider"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Annotate the tracer with the library name. In Jaeger, this shows up as a tag.
var tr = otel.Tracer("cmd/client")

func makeRequest() error {
	ctx, span := tr.Start(
		context.Background(),
		"make request",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	// Auto instrument the http client.
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	// Optionally, add arbitrary values to the context as "baggage"
	ctx = baggage.ContextWithValues(ctx, attribute.String("username", "donuts"))

	// Creating a request with the trace context will allow information about
	// the trace to propagate to the server (implementation detail:
	// via request headers).
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"http://server:8080/hello",
		nil,
	)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Simulate work before the client is done, so you can view
	// different time lengths in the span in telemetry backends.
	time.Sleep(1 * time.Second)

	res, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent(string(body))

	return nil
}

func main() {
	addr := "otel-agent:4317"

	// Initializing the trace provider with the service name makes
	// all traces easy to find by service.
	shutdown := tracer_provider.Initialize(addr, "client")
	defer shutdown()

	// Wait for the server to start (obviously, in a real project, use retry
	// logic rather than sleeping)
	_, span := tr.Start(context.Background(), "wait for server")
	time.Sleep(10 * time.Second)

	// Adding an event is the same as logging a message in the span.
	span.AddEvent("slept for 10 seconds, to wait for server to come up")
	span.End()

	for i := 0; i < 5; i++ {
		if err := makeRequest(); err != nil {
			panic(err)
		}
	}

	// Record an error, just to see what it looks like in the backend
	_, errSpan := tr.Start(context.Background(), "example error")
	defer errSpan.End()

	errSpan.RecordError(errors.New("example error"))

	// Setting the status fails the entire span. In Jaeger, this causes the
	// span to appear red.
	errSpan.SetStatus(codes.Error, "fail entire span")
}
