FROM golang@sha256:04b95d37ab61bd05b6f163383dbd54da4553be2b427b8980a72f778be4edec6b AS builder
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
