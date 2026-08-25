# StoreMesh Product Service

The Product Service owns the product catalog: product identity, SKU, pricing,
currency, lifecycle state, and catalog reads. It does not own inventory,
orders, payments, or user identity.

## Contract status

The initial protobuf contract is in `proto/storemesh/product/v1/product.proto`.
The runtime uses the in-memory catalog when `DATABASE_URL` is unset and the
PostgreSQL repository when it is configured. Apply `migrations/001_products.sql`
before starting the persistent mode. Authentication and production deployment
remain subsequent steps.

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
