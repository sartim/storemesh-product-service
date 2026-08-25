# syntax=docker/dockerfile:1.7

FROM golang:1.26.6-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/storemesh-product-service ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/storemesh-product-service /app/storemesh-product-service
USER nonroot:nonroot
ENTRYPOINT ["/app/storemesh-product-service"]
