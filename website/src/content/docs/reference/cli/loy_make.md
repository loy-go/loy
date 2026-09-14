---
title: "loy make"
description: "Scaffold application components, slices, and vertical features"
slug: reference/cli/loy-make
sidebar:
  order: 7
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

* [loy](/reference/cli/loy/)	 - Loy — Go developer platform with Laravel-like DX
* [loy make ci](/reference/cli/loy-make-ci/)	 - Scaffold CI/CD pipeline workflow (GitHub Actions or GitLab CI)
* [loy make crud](/reference/cli/loy-make-crud/)	 - Scaffold full vertical CRUD slice (migration, queries, model, repo, service, handler, test, wiring)
* [loy make deploy](/reference/cli/loy-make-deploy/)	 - Scaffold production deployment assets (docker, k8s, helm, ci, or all)
* [loy make docker](/reference/cli/loy-make-docker/)	 - Scaffold production multi-stage Dockerfile and docker-compose.yml
* [loy make event](/reference/cli/loy-make-event/)	 - Scaffold domain event struct
* [loy make feature](/reference/cli/loy-make-feature/)	 - Scaffold full vertical feature slice (model, repo, service, handler, test, wiring)
* [loy make handler](/reference/cli/loy-make-handler/)	 - Scaffold HTTP transport handler
* [loy make helm](/reference/cli/loy-make-helm/)	 - Scaffold Helm chart for application (deploy/helm/<name>/)
* [loy make job](/reference/cli/loy-make-job/)	 - Scaffold Asynq background job payload & processor
* [loy make k8s](/reference/cli/loy-make-k8s/)	 - Scaffold cloud-native Kubernetes manifests (deploy/k8s/)
* [loy make listener](/reference/cli/loy-make-listener/)	 - Scaffold event listener consumer
* [loy make model](/reference/cli/loy-make-model/)	 - Scaffold domain entity model
* [loy make policy](/reference/cli/loy-make-policy/)	 - Scaffold authorization policy checks
* [loy make repository](/reference/cli/loy-make-repository/)	 - Scaffold domain repository interface & adapter
* [loy make request](/reference/cli/loy-make-request/)	 - Scaffold HTTP request DTO with validation
* [loy make resource](/reference/cli/loy-make-resource/)	 - Scaffold API response resource transformation
* [loy make runtime](/reference/cli/loy-make-runtime/)	 - Scaffold application runtime lifecycle, composition root, and health checks
* [loy make service](/reference/cli/loy-make-service/)	 - Scaffold application service use case
* [loy make test](/reference/cli/loy-make-test/)	 - Scaffold unit and integration tests
* [loy make view](/reference/cli/loy-make-view/)	 - Scaffold Templ view component or page

