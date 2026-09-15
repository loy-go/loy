---
title: "Migrating from Laravel to Go with Loy"
description: "A complete guide for Laravel developers transitioning to Go: Artisan to Loy CLI, Eloquent to sqlc, and explicit wiring."
---

Laravel developers are accustomed to world-class developer ergonomics: instant scaffolding (`php artisan make:*`), integrated database migrations, scheduled jobs, and unified toolchains. However, scaling dynamic PHP applications often hits memory, concurrency, or type-safety limits.

Loy brings Laravel-grade developer velocity to Go while preserving Go's raw performance, type safety, and zero-runtime footprint.

---

## 1. Concept & Toolchain Rosetta Stone

| Laravel Concept | Loy Equivalent in Go | Distinction in Loy |
|---|---|---|
| `php artisan make:model` | `loy make model <name>` | Generates typed Go domain entity without ActiveRecord magic |
| `php artisan make:migration` | `loy migrate create <name>` | Timestamped standard SQL migration for embedded Goose engine |
| `php artisan make:controller` | `loy make handler <name>` | HTTP transport handler with dependency injection |
| `php artisan route:list` | `loy routes` | Static AST analysis of routes without booting the app |
| `php artisan serve` | `loy dev` | Hot-reloading daemon with process supervision |
| **Eloquent ORM** | **sqlc + pgx/mysql** | Type-safe SQL compiled to Go functions at build time |
| **Service Providers** | `internal/app/wiring.go` | Compile-time explicit Go constructors (zero reflection) |
| **Queues & Jobs** | **Asynq / Valkey** | Typed background jobs with redis-compatible queue runner |
| **Blade Templates** | **Templ + HTMX** | Type-checked HTML components compiled to pure Go |

---

## 2. From Eloquent ActiveRecord to sqlc

Eloquent blends domain entities and database queries into mutable classes:

```php
// Laravel: ActiveRecord (implicit queries, runtime magic)
$user = User::where('email', $email)->first();
$user->status = 'active';
$user->save();
```

In Loy, persistence is strictly separated from domain logic ([ADR-001](/loy/adrs/)):

```sql
-- internal/database/queries/users.sql
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: UpdateUserStatus :exec
UPDATE users SET status = $2, updated_at = NOW() WHERE id = $1;
```

Run `sqlc generate` to compile type-safe Go repository code:

```go
// Loy: Type-safe, compile-time verified database operations
user, err := repo.Queries.GetUserByEmail(ctx, email)
if err != nil {
    return fmt.Errorf("user not found: %w", err)
}
err = repo.Queries.UpdateUserStatus(ctx, db.UpdateUserStatusParams{
    ID:     user.ID,
    Status: "active",
})
```

---

## 3. From Service Providers to Explicit Wiring

Laravel uses dynamic service containers with runtime reflection and magic facades (`Auth::user()`, `Cache::get()`):

```php
// Laravel: Service Container (dynamic resolution)
$this->app->singleton(PaymentGateway::class, function ($app) {
    return new StripePaymentGateway(config('services.stripe.secret'));
});
```

Loy enforces compile-time explicit wiring ([ADR-003](/loy/adrs/)) in `internal/app/wiring.go`:

```go
// internal/app/wiring.go
func (a *App) wireDependencies() error {
    // 1. Repositories
    userRepo, err := userRepo.NewPostgresRepository(a.db)
    if err != nil {
        return err
    }

    // 2. Application Services
    userSvc, err := userService.NewService(userRepo)
    if err != nil {
        return err
    }

    // 3. Handlers
    userHandler, err := userHttp.NewHandler(userSvc)
    if err != nil {
        return err
    }

    // 4. Routes
    userHandler.RegisterRoutes(a.router.Group("/api/v1"))
    return nil
}
```

If a dependency is missing or misconfigured, **the Go compiler catches it at build time**, eliminating runtime container exceptions in production.

---

## 4. Step-by-Step Migration Strategy

1. **Scaffold the App**:
   ```bash
   loy new my-app --preset saas --http fiber --db postgres
   ```
2. **Port Schema Migrations**:
   Run `loy migrate create <table_name>` and write standard PostgreSQL DDL. Apply with `loy migrate up`.
3. **Generate Domain Slices**:
   ```bash
   loy make crud order total:float customer_id:uuid status:string
   ```
4. **Enforce Architecture**:
   Run `loy check` in your CI/CD pipeline to verify layer purity:
   ```bash
   loy check --format github
   ```
