FROM golang:latest@sha256:6f0b0a314b158ff6caf8f12d7f6f3a966500ec6afb533e986eca7375e2f7560f AS builder
WORKDIR /app
COPY go.mod . 
COPY go.sum .
RUN go mod download
COPY . .

FROM builder AS prodBuilder
ARG CMD
RUN CGO_ENABLED=0 go build -ldflags="-w -s" "./cmd/${CMD}"

FROM scratch AS prod
ARG CMD
COPY --from=prodBuilder "/app/${CMD}" "/app/${CMD}"
