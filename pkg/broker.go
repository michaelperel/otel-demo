package pkg

import (
	"context"
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const BrokerKey = "work"

type Broker struct{ c *Client }

func NewBroker(c *Client) *Broker { return &Broker{c: c} }

func (b *Broker) Publish(ctx context.Context, body string) error {
	ctx, span := tr.Start(
		ctx,
		"broker publish",
		// Optionally, annotate the span with a kind. Producer seems to make
		// sense for generically "producing" a message by publishing.
		trace.WithSpanKind(trace.SpanKindProducer),
	)
	defer span.End()

	m := &Message{Body: body, Header: http.Header{}}
	// Since redis' pubsub is not auto instrumented like HTTP,
	// we add http headers to each message and inject the trace
	// context into those headers.
	//
	// When the message is read by the worker service, these headers
	// will be extracted, and spans can be started as children of the
	// current span.
	var pr propagation.TraceContext
	pr.Inject(ctx, propagation.HeaderCarrier(m.Header))

	byt, err := json.Marshal(m)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if err = b.c.Publish(ctx, BrokerKey, byt).Err(); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func (b *Broker) Subscribe(ctx context.Context) (<-chan *Message, error) {
	// This whole method is just boilerplate so that callers can
	// subscribe to a channel of our message type, rather than redis'
	ctx, span := tr.Start(ctx, "broker subscribe")
	defer span.End()

	p := b.c.Subscribe(ctx, BrokerKey)

	// Ensure subscription is created before returning channel
	_, err := p.Receive(ctx)
	if err != nil {
		span.RecordError(err)

		_ = p.Close()

		return nil, err
	}

	pCh := p.Channel()
	mCh := make(chan *Message)

	go func() {
		ctx, span := tr.Start(ctx, "broker listening")
		defer span.End()

		defer close(mCh)
		defer p.Close()

		for {
			select {
			case rawMsg := <-pCh:
				m := &Message{}
				err := json.Unmarshal([]byte(rawMsg.Payload), m)
				if err != nil {
					span.RecordError(err)

					continue
				}

				select {
				case mCh <- m:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return mCh, nil
}
