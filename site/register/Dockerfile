# ---- Build stage ------------------------------------------------------------
FROM golang:1.25 AS builder
WORKDIR /src

# copy mod files first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# copy source and build
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /out/register ./register.go

# ---- Runtime stage ----------------------------------------------------------
FROM alpine:3.20
WORKDIR /app

# copy the compiled binary
COPY --from=builder /out/register /app/register

# expose internal port
EXPOSE 8080

# run directly
ENTRYPOINT ["/app/register"]
