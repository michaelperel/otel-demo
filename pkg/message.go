package pkg

import "net/http"

type Message struct {
	Body string
	// Since redis' pubsub protocol does not have headers like the
	// HTTP protocol, use the trace context to set the same headers that
	// would be in an HTTP request. Specifically, the 'traceparent' header
	// that contains the trace ID and span ID:
	// https://github.com/open-telemetry/opentelemetry-go/blob/d616df61f5d163589228c5ff3be4aa5415f5a884/propagation/trace_context_test.go#L38
	// https://www.w3.org/TR/trace-context/#traceparent-header
	Header http.Header
}
