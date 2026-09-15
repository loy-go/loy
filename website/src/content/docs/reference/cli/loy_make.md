---
title: "loy make"
description: "Scaffold application components, slices, and vertical features"
slug: reference/cli/loy-make
sidebar:
  order: 10
---


Scaffold application components, slices, and vertical features

### Synopsis

The make command tree provides scaffolding for clean-architecture Go applications:
  - Atomic generators: model, repository (repo), service (svc), handler, request, resource, job, event, listener, policy, test
  - Composers: feature, crud

### Options

```
      --dry-run         Preview generated operations without writing to disk
      --force           Overwrite existing files if developer owned
  -h, --help            help for make
      --modular         Scaffold sub-domain modular wiring file instead of flat wiring
      --target string   Target application module in workspace
```

### Options inherited from parent commands

```
  -C, --directory string   Change execution directory
      --json               Output results in JSON format
      --no-color           Disable colored ANSI output
  -q, --quiet              Suppress non-essential output
  -v, --verbose            Enable verbose/debug output
```

### SEE ALSO

* [loy](/loy/reference/cli/loy/)	 - Loy — Go developer platform with Laravel-like DX
* [loy make auth](/loy/reference/cli/loy-make-auth/)	 - Scaffold baseline JWT and password authentication kit
* [loy make ci](/loy/reference/cli/loy-make-ci/)	 - Scaffold CI/CD pipeline workflow (GitHub Actions or GitLab CI)
* [loy make crud](/loy/reference/cli/loy-make-crud/)	 - Scaffold full vertical CRUD slice (migration, queries, model, repo, service, handler, test, wiring)
* [loy make deploy](/loy/reference/cli/loy-make-deploy/)	 - Scaffold production deployment assets (docker, k8s, helm, ci, or all)
* [loy make docker](/loy/reference/cli/loy-make-docker/)	 - Scaffold production multi-stage Dockerfile and docker-compose.yml
* [loy make event](/loy/reference/cli/loy-make-event/)	 - Scaffold domain event struct
* [loy make feature](/loy/reference/cli/loy-make-feature/)	 - Scaffold full vertical feature slice (model, repo, service, handler, test, wiring)
* [loy make grpc](/loy/reference/cli/loy-make-grpc/)	 - Scaffold Proto contract and gRPC transport server
* [loy make handler](/loy/reference/cli/loy-make-handler/)	 - Scaffold HTTP transport handler
* [loy make helm](/loy/reference/cli/loy-make-helm/)	 - Scaffold Helm chart for application (deploy/helm/<name>/)
* [loy make job](/loy/reference/cli/loy-make-job/)	 - Scaffold Asynq background job payload & processor
* [loy make k8s](/loy/reference/cli/loy-make-k8s/)	 - Scaffold cloud-native Kubernetes manifests (deploy/k8s/)
* [loy make listener](/loy/reference/cli/loy-make-listener/)	 - Scaffold event listener consumer
* [loy make model](/loy/reference/cli/loy-make-model/)	 - Scaffold domain entity model
* [loy make outbox](/loy/reference/cli/loy-make-outbox/)	 - Scaffold Transactional Outbox migration, store, and dispatcher
* [loy make policy](/loy/reference/cli/loy-make-policy/)	 - Scaffold authorization policy checks
* [loy make repository](/loy/reference/cli/loy-make-repository/)	 - Scaffold domain repository interface & adapter
* [loy make request](/loy/reference/cli/loy-make-request/)	 - Scaffold HTTP request DTO with validation
* [loy make resource](/loy/reference/cli/loy-make-resource/)	 - Scaffold API response resource transformation
* [loy make runtime](/loy/reference/cli/loy-make-runtime/)	 - Scaffold application runtime lifecycle, composition root, and health checks
* [loy make seeder](/loy/reference/cli/loy-make-seeder/)	 - Scaffold database seeder fixture
* [loy make service](/loy/reference/cli/loy-make-service/)	 - Scaffold application service use case
* [loy make tenant](/loy/reference/cli/loy-make-tenant/)	 - Scaffold multi-tenancy context, RLS helper, and initial migration
* [loy make test](/loy/reference/cli/loy-make-test/)	 - Scaffold unit and integration tests
* [loy make view](/loy/reference/cli/loy-make-view/)	 - Scaffold Templ view component or page
* [loy make ws](/loy/reference/cli/loy-make-ws/)	 - Scaffold WebSocket hub, client pumps, and protocol frames

