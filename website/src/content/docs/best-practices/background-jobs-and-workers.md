---
title: "Best Practices: Background Jobs & Worker Daemons"
description: "How to design resilient background tasks, retries, and worker daemons in Loy."
---

Generated via: `loy make job <name>`

Asynchronous tasks allow applications to offload compute-heavy work, integrate with slow third-party webhooks, and process batch operations without degrading HTTP response times ([ADR-019](/loy/adrs/)).

## Golden Rules

### 1. Minimal Task Payloads
- **DO NOT** embed large binary blobs or full database structs inside background task payloads.
- **DO** store identifiers (`ID`, `UUID`, `OrgID`) and fetch fresh entity state from the database inside the processor:

```go title="internal/cv_parser/job/cv_parse_job.go"
package job

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
)

const TypeCVParseProcess = "cv_parse:process"

type CVParsePayload struct {
	CandidateID int64  `json:"candidate_id"`
	FileKey     string `json:"file_key"`
}

func NewCVParseTask(candidateID int64, fileKey string) (*asynq.Task, error) {
	payload, err := json.Marshal(CVParsePayload{
		CandidateID: candidateID,
		FileKey:     fileKey,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling task payload: %w", err)
	}
	return asynq.NewTask(TypeCVParseProcess, payload, asynq.MaxRetry(5)), nil
}
```

### 2. Idempotent Task Processors
- Network retries and server crashes can cause tasks to be executed more than once.
- Always ensure processors are **idempotent**: running the task multiple times with the same payload must produce the exact same final state without duplicate records or charges.

### 3. Automatic Splicing into Worker Daemon
- When `loy make job <name>` is executed, Loy automatically splices the handler into `internal/app/worker_wiring.go`.
- Both `cmd/api` and `cmd/worker` are supervised simultaneously under `loy dev`.
