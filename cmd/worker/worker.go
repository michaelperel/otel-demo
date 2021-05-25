package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/michaelperel/otel-demo/pkg"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var tr = otel.Tracer("worker")

func main() {
	var (
		otelURL  = mustGetEnvStr("OTEL_AGENT_URL")
		redisURL = mustGetEnvStr("REDIS_URL")
	)

	shutdown := pkg.InitializeGlobalTracer(otelURL, "worker")
	defer shutdown()

	msgs := mustSubscribe(redisURL)

	// listen to messages published to redis' pubsub and perform work
	// based on the message.
	for m := range msgs {
		go work(m)
	}
}

func work(m *pkg.Message) {
	// extract the trace context from the headers of the message
	var pr propagation.TraceContext
	ctx := pr.Extract(
		context.Background(),
		propagation.HeaderCarrier(m.Header),
	)

	_, span := tr.Start(
		ctx,
		"work",
		// Optionally, annotate the span with a kind. Consumer seems to make
		// sense for generically "consuming" a message from a broker.
		trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()

	// simulate work
	time.Sleep(1 * time.Second)
}

func mustGetEnvStr(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic(fmt.Sprintf("'%s' environment variable missing", k))
	}

	return v
}

func mustSubscribe(url string) <-chan *pkg.Message {
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
	msgs, err := b.Subscribe(context.Background())
	if err != nil {
		panic(err)
	}

	return msgs
}
