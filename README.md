## Overview
A small Vercel-style web app for collecting and serving single-elimination bracket submissions: a Svelte + TypeScript frontend and a lightweight Go backend that exposes serverless HTTP handlers and persists bracket/teams/predictions to PostgreSQL. It targets quick deploys (Vercel) and uses Protocol Buffers as the canonical message format between codegen targets.

### Stack
- **Language(s):** Svelte (frontend), Go (backend), TypeScript (frontend tooling)
- **Framework / runtime:** Vite + Svelte (frontend); Go 1.26 as serverless handlers (Vercel-compatible)
- **Notable libraries / technologies:** github.com/jackc/pgx/v5 (Postgres driver), google.golang.org/protobuf (protobufs), ts-proto / protobufjs (TS protobuf tooling), vite + @sveltejs/vite-plugin-svelte

## How it's organized
```
api/            -> serverless entrypoint (api/index.go) registering HTTP handlers
server/         -> Go application logic and DB handlers (fetch_bracket.go, submit_bracket.go, delete_bracket.go, submit_teams.go, app.go)
  proto/        -> server-side proto outputs / generated artifacts
proto/          -> top-level .proto definitions (source of truth for message shapes; codegen target)
web/            -> Svelte + TypeScript frontend (vite) with its own package.json and dev/build scripts
scripts/        -> Go helper binary that initializes the DB schema (scripts/main.go + scripts/schema.go)
package.json    -> root helper script: build the web app (runs cd web && npm install && npm run build)
go.mod          -> Go module + dependencies
vercel.json     -> Vercel/serverless configuration
```

How it fits together:
- The web/ SPA (Vite + Svelte) is the user-facing app and calls REST endpoints under /api/*.
- api/index.go is the serverless function entrypoint: it reads DATABASE_URL, opens a pgx pool, creates server.App and registers handlers (GET/POST/DELETE /api/bracket, POST /api/teams).
- server/ implements the business logic and SQL interactions for matches, teams, users, and predictions.
- Protocol buffers in proto/ are used as the canonical message/schema definition and are code-generated for Go (server) and TypeScript (web) via ts-proto / protobufjs tooling.
- scripts/ contains a small Go program (scripts/main.go) which uses godotenv + pgx to create the database schema (scripts/schema.go) — this is the repository's DB initializer.

## How to run it
Frontend (dev)
```
cd web
npm install
npm run dev
# vite dev server 
```

Build frontend (from repo root)
```
npm run build
# runs: cd web && npm install && npm run build
```

Initialize database (local)
```
# from repo root, build/run scripts/main.go or run with `go run`
go run ./scripts main.go
# or:
go run ./scripts
# Requires a .env or environment variable:
DATABASE_URL=postgres://user:pass@host:5432/dbname
```

Deploy serverless
- Deploy to Vercel: api/index.go will be used as a Go serverless handler.
- Required environment variable:
  - DATABASE_URL — Postgres connection string (postgres://user:pass@host:5432/db)

Codegen (proto)
- Use protoc + protoc-gen-go for Go, and ts-proto / protobufjs for TypeScript. Example (adjust paths/plugins):
```
protoc --go_out=. --go_opt=paths=source_relative proto/*.proto
protoc --plugin=protoc-gen-ts_proto=./node_modules/.bin/protoc-gen-ts_proto \
  --ts_proto_out=web/src/lib/proto \
  proto/*.proto
```
