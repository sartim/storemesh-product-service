# StoreMesh Product Service

The Product Service owns the product catalog: product identity, SKU, pricing,
currency, lifecycle state, and catalog reads. It does not own inventory,
orders, payments, or user identity.

## Contract status

The initial protobuf contract is in `proto/storemesh/product/v1/product.proto`.
The runtime uses the in-memory catalog when `DATABASE_URL` is unset and the
PostgreSQL repository when it is configured. Apply `migrations/001_products.sql`
before starting the persistent mode. Authentication and production deployment
are enabled with `JWT_SECRET`, `JWT_ISSUER`, and `JWT_AUDIENCE`; when
`JWT_SECRET` is unset, the binary is intentionally unauthenticated for local
contract development only.

Keycloak access tokens can also be accepted during migration by setting
`KEYCLOAK_ISSUER` and `KEYCLOAK_AUDIENCE`. The service discovers the issuer's
JWKS endpoint and validates RS256 tokens. Legacy HS256 service tokens remain
supported while downstream services migrate; do not expose unauthenticated
mode outside isolated local development.

## Local validation

Install [Buf](https://buf.build) and run:

```sh
buf lint proto
buf build proto
cd proto && buf generate --template buf.gen.yaml .
```

Generation writes Go/protobuf transport code to `gen/` and the OpenAPI
document to `openapi/`. These are build artifacts for the upcoming runtime
implementation and should be regenerated whenever the contract changes.

The service will follow the StoreMesh conventions: gRPC internally, annotated
HTTP handlers externally, PostgreSQL persistence, OpenTelemetry, Prometheus,
structured logging, and Helm/Argo CD delivery.

## Run locally without Docker or Kubernetes

Requires Go 1.26.6 or newer. The default in-memory catalog needs no external
services:

```sh
go run ./cmd/server
```

Use alternate addresses when running multiple services directly on localhost:

```sh
GRPC_ADDR=:50051 METRICS_ADDR=:8081 go run ./cmd/server
```

Set `DATABASE_URL` and apply `migrations/001_products.sql` to exercise the
PostgreSQL repository. Set `JWT_SECRET` only when testing authenticated gRPC;
an unset secret intentionally keeps local contract development unauthenticated.
