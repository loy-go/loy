# Phase 9: Fullstack Integration Implementation Plan

**Phase:** 9 of 10  
**Status:** Completed  
**Estimated Scope:** Templ SSR component scaffolding, HTMX patterns, Vite asset pipeline  
**Primary Specifications:** [01-PRD.md](../01-PRD.md), [07-Integration-Capability-Specification.md](../07-Integration-Capability-Specification.md), [15-Development-Workflow-Toolchain.md](../15-Development-Workflow-Toolchain.md), [ADR-013](../adrs/ADR-013-codegen-engine-selection.md)

---

## 1. Goal & Objectives
Provide a modern, lightweight fullstack Go web development experience:
- Server-Side Rendering (SSR) via `a-h/templ` components.
- Dynamic frontend updates using HTMX conventions (swaps, partials, out-of-band updates).
- Asset bundling orchestration with Vite (Tailwind CSS, TypeScript, HMR) without replacing standard JS package managers.
- Fullstack project preset integration (`loy new <app> --preset fullstack`).

---

## 2. Package Architecture of Fullstack Scaffold

```text
<generated_fullstack_app>/
├── cmd/
│   └── web/
│       └── main.go         # Web application entrypoint
├── internal/
│   └── transport/
│       └── http/
│           ├── handler/    # Web controllers rendering Templ views
│           └── static/     # Static file embedding (embed.FS for production)
├── views/
│   ├── layouts/
│   │   └── base.templ      # HTML skeleton, Tailwind inclusion, HTMX script tag
│   ├── components/
│   │   ├── navbar.templ    # Reusable UI widgets
│   │   └── alert.templ
│   └── pages/
│       └── home.templ      # Page templates
├── assets/
│   ├── css/
│   │   └── app.css         # Tailwind directives
│   └── js/
│       └── app.ts          # HTMX extensions & client interactivity
├── package.json            # Vite, Tailwind CSS, Prettier
├── vite.config.ts          # Vite build config outputting to dist/ or static/
└── loy.yaml                # assets: vite, template: templ
```

---

## 3. Concrete Implementation Steps

### Step 9.1: Templ SSR Conventions
1. Scaffold base layout `views/layouts/base.templ`:
   - Viewport meta, CSRF token header, HTMX CDN or local bundle.
   - Dynamic body slot for child components.
2. Provide Fiber adapter for Templ rendering:
   ```go
   func Render(c *fiber.Ctx, component templ.Component) error {
       c.Set("Content-Type", "text/html; charset=utf-8")
       return component.Render(c.Context(), c.Response().BodyWriter())
   }
   ```

### Step 9.2: HTMX Scaffolding & Patterns
1. Support partial view generation:
   - Provide command `loy make view <name> [--partial]`.
   - Include HTMX attributes (`hx-get`, `hx-post`, `hx-target`, `hx-swap`).
2. Add HTMX helper middleware in Fiber to detect `HX-Request` headers and render full page vs partial layout accordingly.

### Step 9.3: Vite Toolchain Orchestration
1. Generate `package.json` and `vite.config.ts`:
   - Configured for Tailwind CSS v3/v4.
   - Outputs compiled bundle to `internal/transport/http/static/dist/`.
2. Connect Vite dev server into `loy dev`:
   - When running `loy dev`, supervisor spawns `npm run dev` in parallel with Go server.
   - Proxy asset requests or inject Vite HMR client during development.
   - Embed compiled assets via `//go:embed static/dist/*` in production build.

---

## 4. Test Strategy & Acceptance Criteria

### SSR Rendering Tests
- Call handler with Templ view -> verify HTTP response returns valid HTML string with 200 OK.
- Query with header `HX-Request: true` -> verify only partial component renders without layout shell.

### Asset Bundling Tests
- Run `npm run build` -> verify `dist/manifest.json` and bundled CSS/JS generated.
- Run `go build ./...` -> verify binary embeds assets without disk dependency.

---

## 5. Definition of Done
- [x] Fullstack preset generates operational Fiber + Templ + HTMX + Vite app.
- [x] Live reload functions seamlessly between Templ edits, CSS changes, and Go logic.
- [x] Production build produces single standalone binary containing embedded web assets.

---

[← Previous: Phase 8 Plan](./08-Phase-Developer-Experience.md) | [Back to Plans Index](./README.md) | [Next: Phase 10 Plan →](./10-Phase-Deployment.md)
