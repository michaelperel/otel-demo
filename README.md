# What is this?
This is a demo of Open Telemetry's distributed tracing capabilities.
In `docker-compose.yml` there are variety of services:
* `server` - a service that implements an http server
* `client` - a service that sends a few requests to the server
* `jaeger` - an open source telemetry backend
* `zipkin` - an open source telemetry backend
* `otel-agent` - a service that receives traces from `server` and `client`
* `otel-collector` - a service that receives traces forwarded from `otel-agent`
  and exports them to `jaeger` and `zipkin`

# Why is this interesting?
By using Open Telemetry with the agent and collector, backends are swappable
and all services handle tracing in the same way, regardless of language.

Specifically, applications send traces to the agent, which forwards them to the
collector, and the collector defines backends via exporters in yaml.

Here we use 2 exporters, `jaeger` and `zipkin`, but there are many possible
exporters including
[Azure Monitor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/azuremonitorexporter).

# How to use?
`docker-compose up --build` brings up all services. 
The `client` sends a few requests to `server`. The distributed traces appear in
`jaeger` and `zipkin`.

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
Start by reading the comments in `src/cmd/client/client.go`.
They describe how to create a trace that propagates to the server over
an http request.

Next, read the comments in `src/cmd/server/server.go`. They describe
how the propagated trace is used in subsequent spans.

Finally, read the comments in `src/pkg/tracer_provider/tracer_provider.go`. They
describe boilerplate code that sets up a tracer provider for each application.

# Development
A dev container has been provided. To use:
* Ensure the `Remote - Containers` extension is installed in VSCode
* Open the project in the container
* Install the Go extension libraries with `Go: Install/Update tools` from
  the command palette 

# Citations
The collector code is adapted from
[this official otel example](https://github.com/open-telemetry/opentelemetry-collector/tree/main/examples/demo).

The client / server code is adapted from 
[this official otel example](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/net/http/otelhttp/example).
