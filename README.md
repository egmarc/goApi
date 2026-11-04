# goApi

A Docker-ready REST API using Go's standard-library `net/http` router.
Dependencies are managed with Go modules; no third-party packages are needed.

## Layout

```text
cmd/api/            Application entry point, configuration, and shutdown
internal/api/       HTTP routes and route-level authorization middleware
tests/integration/  Integration tests using a real HTTP test server
Dockerfile          Multi-stage build with a minimal, non-root runtime
docker-compose.yml  Local container configuration
```

## Run locally

Requires Go 1.24.13 or newer, or Docker with Docker Compose.

Generate a token and keep it in your environment (never commit it):

```sh
export API_TOKEN="$(openssl rand -hex 32)"
go run ./cmd/api
```

`API_TOKEN` is required and must not contain whitespace. `HTTP_ADDR` defaults
to `:8080`; for host-only development, set it to `127.0.0.1:8080`.

Alternatively, with the same `API_TOKEN` environment variable:

```sh
docker compose up --build -d
docker compose logs api
docker compose down
```

Compose also accepts `API_TOKEN` from a local, gitignored `.env` file. It binds
the API to localhost port 8080 and runs it as a non-root user with a read-only
filesystem. The Docker build uses the current Go 1 Alpine image.

## Routes and authorization

| Method | Route | Authorization | Response |
| --- | --- | --- | --- |
| GET | `/healthz` | Public | `{"status":"ok"}` |
| GET | `/api/v1/protected` | Required (`API_TOKEN`) | `{"message":"authorized"}` |

```sh
curl http://localhost:8080/healthz
curl -i http://localhost:8080/api/v1/protected
curl --oauth2-bearer "$API_TOKEN" http://localhost:8080/api/v1/protected
```

Missing or invalid credentials return HTTP 401 with a bearer authentication
challenge. Unknown routes return 404; unsupported methods return 405.
Go's router also supports HEAD for GET routes, retaining authorization.

Register new routes in `internal/api/handler.go`. Wrap each restricted handler
with `requireToken` as shown by `/api/v1/protected`; public handlers are
registered directly. This starter uses one shared token, not user identities,
roles, JWTs, or an OAuth provider. Use HTTPS via a reverse proxy before exposing
the API outside local development; bearer tokens must not travel over public
unencrypted connections. Do not log authorization headers.

The server has HTTP timeouts and gracefully shuts down on SIGINT or SIGTERM.

## Validate

```sh
go test ./tests/integration -v
go test -race ./...
go vet ./...
go build -o bin/api ./cmd/api
```

Integration tests start an isolated server with `httptest`, exercising the same
router used by the application over HTTP. They need neither Docker nor a
running API and cover public routes, authorized requests, malformed credentials,
and route/method handling.
