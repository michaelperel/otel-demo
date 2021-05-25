package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/michaelperel/otel-demo/pkg"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
)

// Annotate the tracer with the library name. In Jaeger, this shows up as a tag.
var tr = otel.Tracer("client")

func main() {
	var (
		otelURL   = mustGetEnvStr("OTEL_AGENT_URL")
		serverURL = mustGetEnvStr("SERVER_URL")
	)

	// Initializing the tracer with the service name makes
	// all traces easy to find by service.
	shutdown := pkg.InitializeGlobalTracer(otelURL, "client")
	defer shutdown()

	// Wait for the server to start (obviously, in a real project, use retry
	// logic rather than sleeping)
	//
	// Start a span. A trace is a collection of spans, where spans can have
	// children spans. Information necessary to create a child span is
	// returned in the first value (normally named "ctx", but we are not
	// creating a child span, so we have assigned it to "_").
	_, span := tr.Start(context.Background(), "wait for server")
	time.Sleep(10 * time.Second)

	// Adding an event is the same as logging a message in the span.
	span.AddEvent("slept for 10 seconds, to wait for server to come up")
	span.End()

	for i := 0; i < 5; i++ {
		if err := makeRequest(serverURL); err != nil {
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

func mustGetEnvStr(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic(fmt.Sprintf("'%s' environment variable missing", k))
	}

	return v
}

func makeRequest(url string) error {
	ctx, span := tr.Start(
		context.Background(),
		"make request",
	)
	defer span.End()

	// Auto instrument the http client. This will create a new span whenever
	// a request is made, and set the span kind to client (setting span kinds,
	// like many attributes, is optional and appears as tags in Jaeger).
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
		url,
		nil,
	)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Simulate work before the client is done, so you can more easily view
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
