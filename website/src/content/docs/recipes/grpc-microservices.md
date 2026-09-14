---
title: "Recipe: gRPC & REST Dual-Listener Microservices"
description: "How to build high-performance internal RPC and public REST endpoints from a single Go binary."
---

Microservice architectures often require fast binary Protobuf RPC for inter-service communication and JSON REST for public API consumers. Loy supports dual-transport services without domain model leakage ([ADR-021](/loy/adrs/)).

## 1. Clean Architecture Protocol Isolation

```
            ┌──► internal/<pkg>/transport/http/ (JSON REST Handlers) ──┐
cmd/api ────┤                                                          ├──► internal/<pkg>/service/ (Domain Logic)
            └──► internal/<pkg>/transport/grpc/ (Proto RPC Handlers) ──┘
```

## 2. Scaffolding gRPC Contracts & Adapters

Run `loy make grpc` to scaffold a Protocol Buffer contract and server adapter:

```bash
loy make grpc billing "amount:int,currency:string,customer_id:string"
```

Scaffolded artifacts:
- **`proto/billing/v1/billing.proto`**: Contract-first Protocol Buffer specification.
- **`internal/billing/transport/grpc/server.go`**: Server adapter calling `billing.Service`.

```proto title="proto/billing/v1/billing.proto"
syntax = "proto3";

package billing.v1;

option go_package = "myapp/proto/billing/v1;billingv1";

service BillingService {
  rpc GetBilling (GetBillingRequest) returns (GetBillingResponse);
  rpc CreateBilling (CreateBillingRequest) returns (CreateBillingResponse);
}
```

## 3. Server Adapter Implementation

The gRPC server adapter converts Proto DTOs into Domain Entities and invokes application services:

```go title="internal/billing/transport/grpc/server.go"
package grpc

import (
	"context"
	"myapp/internal/billing/service"
)

type Server struct {
	service *service.Service
}

func NewServer(svc *service.Service) (*Server, error) {
	return &Server{service: svc}, nil
}

func (s *Server) GetBilling(ctx context.Context, id int64) (*domain.Billing, error) {
	return s.service.GetByID(ctx, id)
}
```

## 4. Dual-Port Server Setup

Run Fiber REST on `:8080` and gRPC on `:9090` under a unified shutdown coordinator:

```go title="cmd/api/main.go"
// Fiber HTTP Listener
go func() {
    _ = app.Router().Listen(":8080")
}()

// gRPC Listener
lis, _ := net.Listen("tcp", ":9090")
grpcServer := grpc.NewServer()
pb.RegisterBillingServiceServer(grpcServer, billingGRPCServer)
go func() {
    _ = grpcServer.Serve(lis)
}()
```
