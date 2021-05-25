package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/michaelperel/otel-demo/pkg"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

var tr = otel.Tracer("server")

func main() {
	var (
		otelURL  = mustGetEnvStr("OTEL_AGENT_URL")
		redisURL = mustGetEnvStr("REDIS_URL")
	)

	shutdown := pkg.InitializeGlobalTracer(otelURL, "server")
	defer shutdown()

	server := mustNewServer(redisURL)

	// Auto instrument any request to /hello. This will create a span whenever
	// a request to this endpoint is made. It will also set the span kind to
	// server.
	otelHandler := otelhttp.NewHandler(server, "/hello")
	http.Handle("/hello", otelHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

type server struct{ b *pkg.Broker }

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Simulate work to make viewing the span easier in the telemetry backend.
	time.Sleep(1 * time.Second)

	// Create a child span, using the trace context propagated from the
	// client request in r.Context().
	ctx, span := tr.Start(
		r.Context(),
		"handle request",
	)
	defer span.End()

	// Baggage from client propagates as well
	uk := attribute.Key("username")
	username := baggage.Value(ctx, uk)
	span.AddEvent(
		"reading username baggage...",
		trace.WithAttributes(uk.String(username.AsString())),
	)

	// Add the response as an attribute to the span. This can show up
	// differently in different backends, but in Jaeger attributes show up
	// as tags.
	body := "hello, world"
	span.SetAttributes(attribute.String("body", body))

	// Publish the body to redis pubsub so that it can be picked up by the
	// worker service
	if err := s.b.Publish(ctx, body); err != nil {
		span.RecordError(err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, body)
}

func mustGetEnvStr(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic(fmt.Sprintf("'%s' environment variable missing", k))
	}

	return v
}

func mustNewServer(url string) *server {
	c, err := pkg.NewClient(url)
	if err != nil {
		panic(err)
	}

	a := time.After(60 * time.Second)
	t := time.NewTicker(3 * time.Second)
loop:
	for {
		select {
		case <-t.C:
			if err := c.Ping(context.Background()).Err(); err == nil {
				break loop
			}
		case <-a:
			panic("timeout connecting to redis")
		}
	}

	b := pkg.NewBroker(c)
	return &server{b}
}
