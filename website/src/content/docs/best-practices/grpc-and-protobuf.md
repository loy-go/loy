---
title: "Best Practices: gRPC & Protocol Buffers"
description: "How to design contract-first Protocol Buffers and clean architecture gRPC server adapters in Loy."
---

Generated via: `loy make grpc <name>` ([ADR-021](/loy/adrs/))

gRPC provides high-performance binary RPC for microservices and inter-service communication.

## Golden Rules

### 1. Proto Contracts are Transport DTOs
- **DO NOT** use generated Protobuf structs as domain entities. Protobuf structs contain internal mutexes, getter methods, and serialization state flags.
- **DO** map Proto messages to domain entities inside the gRPC server adapter:

```go title="internal/billing/transport/grpc/server.go"
package grpc

import (
	"context"
	pb "myapp/proto/billing/v1"
	"myapp/internal/billing/service"
)

type Server struct {
	pb.UnimplementedBillingServiceServer
	service *service.Service
}

func (s *Server) GetBilling(ctx context.Context, req *pb.GetBillingRequest) (*pb.GetBillingResponse, error) {
	entity, err := s.service.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.GetBillingResponse{
		Billing: &pb.Billing{
			Id:     entity.ID,
			Amount: int32(entity.AmountCents),
		},
	}, nil
}
```

### 2. Contract-First Toolchain with `buf`
- Manage `.proto` files inside `proto/<domain>/v1/`.
- Use `buf` for automated linting, breaking change detection, and Go code generation.
