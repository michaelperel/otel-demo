# What is this?
This is a demo of Open Telemetry's distributed tracing capabilities.
In `docker-compose.yml` there are variety of services:
* `client` - a service that sends a few requests to the server
* `server` - a service that implements an HTTP server and publishes a message
  per request via [redis' pubsub](https://redis.io/topics/pubsub)
* `worker` - a service that listens for messages on redis' pubsub and
  does work when a message is published
* `jaeger` - an open source telemetry backend
* `zipkin` - an open source telemetry backend
* `otel-agent` - a service that receives traces from `server` and `client`
* `otel-collector` - a service that receives traces forwarded from `otel-agent`
  and exports them to `jaeger` and `zipkin`

![Architecture](./docs/architecture.png)

# Why is this interesting?
1. By using Open Telemetry with the collector, backends are swappable
   and all services handle tracing in the same way, regardless of programming
   language.

   Specifically, applications send traces to the agent, which forwards them to
   the collector, and the collector defines backends via exporters in yaml.

   Here we use 2 exporters, `jaeger` and `zipkin`, but there are many possible
   exporters including
   [Azure Monitor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/azuremonitorexporter).

2. Cloud architectures often use some form of a message broker to communicate
   long running operations. While HTTP is covered via docs, many messaging
   systems use protocols that are not supported by the Open Telemetry SDK
   (there are no helper functions that inject and extract spans for you).
   One such example would be redis' pubsub wire protocol. In this repo, we show
   how to add distributed tracing to any arbitrary messaging system.

3. Many popular libraries integrate with Open Telemetry with no extra work
   required. One library is [go-redis](https://github.com/go-redis/redis). This
   is great because if a library is not instrumented, the best you can do is
   either modify the library or instrument code which calls the library (which
   inherently misses internal events in the library that do not bubble up to
   the surface of the exposed API).

# How to use?
`docker-compose up --build` brings up all services. 

The `client` sends a few requests to `server`. The `server` publishes messages
to `redis`. The `worker` listens for messages and performs work when they are
published.

The distributed traces appear in `jaeger` and `zipkin`.

`jaeger` can be accessed at `http://localhost:16686`.

`zipkin` can be accessed at `http://localhost:9411`.

`docker-compose down` cleans up all resources.

If you would like to manually make requests to the server after the client ends,
navigate to `http://localhost:8080/hello`.

After requests have been made, if you choose the `client` service in `jaeger`,
you should see something similar to:

![Overview](./docs/jaeger.png)

Note that you can see all traces that started from the client. If you click on
a trace, you can see the distributed spans that make up the trace:

![Spans](./docs/jaeger-span.png)

# How to navigate the code?
Start by reading the comments in `cmd/client/client.go`.
They describe how to create a trace that propagates to the server via
an HTTP request.

Next, read the comments in `cmd/server/server.go`. They describe
how the propagated trace is used in children spans.

Next, read the comments in `pkg/message.go`. They describe how to
add headers to the message that propagate the trace context from the `server`
to the `worker`, in the same way as would be done via HTTP.

Next, read the comments in `cmd/worker/worker.go`. They describe how to
extract the trace context from messages on redis' pubsub and create child spans
with this context.

Next, read the comments in `pkg/broker.go`. They describe how the trace context
can be manually injected and extracted, when publishing and receiving messages.

Finally, read the comments in `pkg/tracer.go`. They describe boilerplate code
that sets up a tracer provider for each application.

# Development
A dev container has been provided. To use:
* Ensure the `Remote - Containers` extension is installed in VSCode
* Open the project in the container
* Install the Go extension libraries with `Go: Install/Update tools` from
  the command palette

> Note: When running any docker commands, run them from outside of the
dev container (on the host machine)

# Citations
The collector code is adapted from
[this official otel example](https://github.com/open-telemetry/opentelemetry-collector/tree/main/examples/demo).

The client / server code is adapted from 
[this official otel example](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/net/http/otelhttp/example).
