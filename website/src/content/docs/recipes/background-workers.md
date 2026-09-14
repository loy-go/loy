---
title: "Recipe: Background Worker Fleet"
description: "How to operate dual-daemon background task workers supervised alongside your HTTP API."
---

Heavy background tasks (PDF generation, video transcode, AI scoring, email delivery) should never run inside the user-facing HTTP process. Loy scaffolds a **Dual-Daemon Topography** (`cmd/api` and `cmd/worker`) supervised together under `loy dev` ([ADR-019](/loy/adrs/)).

## 1. Topography Overview

```
                        ┌──► cmd/api/main.go (HTTP/REST & WebSocket Server)
loy dev (Supervisor) ───┤
                        └──► cmd/worker/main.go (Asynq / River Task Consumer)
```

## 2. Scaffolding Worker Daemons

Create a new project with worker daemon support:

```bash
loy new job-runner --preset=api --queue=asynq
cd job-runner
```

Scaffolded directory layout:

```
job-runner/
├── cmd/
│   ├── api/
│   │   └── main.go              # HTTP server entrypoint
│   └── worker/
│       └── main.go              # Background consumer entrypoint
└── internal/
    └── app/
        ├── app.go               # Base application lifecycle
        ├── wiring.go            # HTTP dependency composition root
        └── worker_wiring.go     # Task registration composition root
```

## 3. Creating and Splicing Background Tasks

Run `loy make job` to scaffold a new background task:

```bash
loy make job pdf_render
```

Loy generates the task payload and processor in `internal/pdf_render/job/` and automatically splices task handling into `internal/app/worker_wiring.go`:

```go title="internal/app/worker_wiring.go"
package app

import (
	"github.com/hibiken/asynq"
	// loy:region:worker_imports
	pdfRenderJob "job-runner/internal/pdf_render/job"
	// loy:endregion
)

// WireWorkerTasks registers all background job handlers with the Asynq ServeMux.
func (a *App) WireWorkerTasks(mux *asynq.ServeMux) error {
	// loy:region:tasks
	mux.HandleFunc(pdfRenderJob.TypePdfRenderProcess, pdfRenderJob.NewPdfRenderProcessor().ProcessTask)
	// loy:endregion

	return nil
}
```

## 4. Enqueuing Tasks in Handlers

Inject the Asynq client into your application service to enqueue background tasks:

```go
func (s *Service) RequestPDF(ctx context.Context, documentID int64) error {
    task, err := pdfRenderJob.NewPdfRenderTask(documentID)
    if err != nil {
        return err
    }

    _, err = s.queueClient.EnqueueContext(ctx, task)
    return err
}
```

## 5. Live Development & Supervision

Start both the API server and worker daemon simultaneously:

```bash
loy dev
```

The terminal supervisor displays multiplexed, colorized output:

```
14:02:01 [api]    listening on :8080
14:02:01 [worker] asynq: server starting concurrency=10
```
