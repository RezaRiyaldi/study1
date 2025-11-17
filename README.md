# Study1

Simple Go project (Gin + Gorm) scaffold with docs generation (swag) and a PowerShell hot-reload helper.

## Prerequisites

- Go 1.20+ installed and `GOPATH`/`%USERPROFILE%\go\bin` available in `PATH`.
- (Optional) `swag` CLI for OpenAPI generation: `go install github.com/swaggo/swag/cmd/swag@latest`.
- (Optional) `CompileDaemon` or `air` for hot reload:
  - `go install github.com/githubnemo/CompileDaemon@latest`
  - or `go install github.com/cosmtrek/air@latest`

## Quick start (Windows PowerShell)

1. Install Go tools (if needed):

```powershell
# swag (OpenAPI generator)
go install github.com/swaggo/swag/cmd/swag@latest

# optional hot-reload tools
go install github.com/githubnemo/CompileDaemon@latest
# or
go install github.com/cosmtrek/air@latest
```

Make sure `%USERPROFILE%\go\bin` is on your `PATH` (restart PowerShell if you add it):

```powershell
setx PATH "$env:PATH;${env:USERPROFILE}\go\bin"
```

2. Allow running local scripts for this session (if needed):

```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
```

3. Generate Swagger docs (from repo root):

```powershell
# generate docs into docs/ (swag must be installed)
& "$env:USERPROFILE\go\bin\swag.exe" init -g cmd/api/main.go --parseDependency --parseInternal --ot go
```

4. Run the API:

```powershell
# run the api server
go run cmd/api/main.go
```

Open the Swagger UI at:

```
http://localhost:8080/swagger/index.html
```

> The `@BasePath` is `/api/v1` (configured in `cmd/api/main.go`), so generated paths will appear under that base.

## Hot reload helper

A PowerShell helper is included at `hot-reload.ps1`.

- If `CompileDaemon` is installed it will use it (recommended). Otherwise it falls back to a FileSystemWatcher that restarts `go run main.go` on `.go` file changes.

Usage (PowerShell, from project root):

```powershell
# auto-detect CompileDaemon (recommended)
.\hot-reload.ps1

# force using CompileDaemon
.\hot-reload.ps1 -UseCompileDaemon
```

## Docs generation notes

- API metadata and base path are in `cmd/api/main.go` (swag uses these comments when generating docs).
- Endpoint annotations are present in handlers and doc-only helper functions in `internal/core/http/server.go` so `swag` picks up root and health endpoints.
- Generated docs live under `docs/` (auto-created by `swag init`).

## Project layout (important files)

- `cmd/api/main.go` — application entry and Swagger meta comments.
- `internal/core/http/server.go` — Gin server, Swagger route, and API root/health/version handlers.
- `internal/modules/user/*` — example module: `handler.go`, `model.go`, `dto.go` (annotated for swag).
- `hot-reload.ps1` — PowerShell watcher/helper for hot reload.
- `docs/` — generated OpenAPI docs from `swag`.

## Environment variables

You can configure via environment variables (see `internal/core/config/config.go`):

- `SERVER_PORT` (default `8080`)
- `SERVER_ENV` (default `development`)
- `DB_HOST`, `DB_NAME`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`

## Suggestions / Next steps

- Add concrete response wrapper types for Swagger so `data` in responses shows precise schemas (e.g., `ListUsersResponse`).
- Add authentication middleware and annotate protected endpoints.
- Add unit tests and CI steps to regenerate docs.

---

If you want, I can:
- Run `swag init` for you and commit the generated `docs/` files.
- Add example `ListUsersResponse` wrapper types and update annotations for more precise Swagger schemas.
- Commit these README and helper files to git.
