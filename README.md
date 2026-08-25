# StoreMesh Product Service

The Product Service owns the product catalog: product identity, SKU, pricing,
currency, lifecycle state, and catalog reads. It does not own inventory,
orders, payments, or user identity.

## Contract status

The initial protobuf contract is in `proto/product/v1/product.proto`. Generated
transport code and persistence are deliberately the next implementation steps;
this repository establishes the boundary before runtime code is added.

## Local validation

Install [Buf](https://buf.build) and run:

```sh
buf lint proto
buf breaking --against '.git#branch=main' proto
```

The service will follow the StoreMesh conventions: gRPC internally, annotated
HTTP handlers externally, PostgreSQL persistence, OpenTelemetry, Prometheus,
structured logging, and Helm/Argo CD delivery.
