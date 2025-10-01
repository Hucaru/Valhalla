FROM golang:1.25 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
RUN go build -o /out/Valhalla

FROM alpine:3.20
WORKDIR /app

COPY drops.json /app/drops.json
COPY reactors.json /app/reactors.json
COPY reactor_drops.json /app/reactor_drops.json
COPY scripts/ /app/scripts/

COPY --from=builder /out/Valhalla /app/Valhalla

RUN echo "#!/bin/sh" > docker-entrypoint.sh && \
  echo "set -ex" >> docker-entrypoint.sh && \
  echo 'exec "$@"' >> docker-entrypoint.sh && \
  chmod +x docker-entrypoint.sh

ENTRYPOINT ["/app/docker-entrypoint.sh"]