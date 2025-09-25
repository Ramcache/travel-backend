# Travel Backend

The Travel Backend is a Go service that powers a travel booking platform. It exposes RESTful endpoints for authentication, trip discovery, orders, news, and profile management, and integrates with PostgreSQL for persistence.

## Tech stack
- Go 1.24
- Chi HTTP router with CORS middleware
- PostgreSQL with Goose migrations
- JWT-based authentication
- Zap structured logging and Prometheus metrics

## Project layout
- `cmd/api` – entry point that wires CLI commands such as `serve` and `migrate`.
- `internal/app` – application wiring and dependency graph composition.
- `internal/server` – HTTP router, middleware, and server bootstrap.
- `internal/handlers` – request/response handlers for auth, users, trips, news, profiles, stats, and orders.
- `internal/services` – domain logic for trips, news, authentication, and supporting features.
- `internal/repository` – PostgreSQL repositories used by the services.
- `internal/storage` – database connection pooling helpers.
- `internal/config` – environment-based configuration loader.
- `migrations` – SQL migrations executed through Goose.

## Prerequisites
- Go >= 1.24.5
- PostgreSQL 13+
- Access to a Telegram bot token and chat ID if you plan to enable Telegram notifications.

## Configuration
Create a `.env` file (or export environment variables in your shell) with the following keys:

| Variable | Description | Default |
| --- | --- | --- |
| `APP_ENV` | Environment name used for logging (`dev`/`prod`). | `dev` |
| `APP_PORT` | HTTP port the API server binds to. | `8080` |
| `APP_JWT_SECRET` | Secret string used to sign JWT tokens. | `changeme` |
| `JWT_TTL` | Token lifetime as Go duration (e.g. `24h`). | `24h` |
| `FRONTEND_URL` | Optional frontend base URL used in notifications. | empty |
| `DB_URL` | PostgreSQL connection string. | empty |
| `DB_MAX_CONNS` | Maximum pooled connections. | `10` |
| `DB_MIN_CONNS` | Minimum pooled connections. | `2` |
| `DB_CONN_TIMEOUT` | Connection acquisition timeout (Go duration). | `5s` |
| `DB_IDLE_TIMEOUT` | Idle connection lifetime (Go duration). | `5m` |
| `TG_TOKEN` | Telegram bot token. | empty |
| `TG_CHAT` | Telegram chat ID for alerts. | empty |

All configuration values are loaded on startup by `internal/config`. When the `.env` file is missing the service falls back to the host environment variables.

## Running the API locally
1. Install dependencies: `go mod download`
2. Export the required environment variables (see above).
3. Run migrations to set up the database schema (see next section).
4. Start the API server:
   ```bash
   go run ./cmd/api serve


You can also build a binary with `task build` and start it via `./bin/travel-api serve`.

The `serve` command wires configuration, logging, database connectivity, and the HTTP router before starting the server on the configured port.

## Database migrations

Database migrations are managed with Goose. The project exposes CLI commands and Taskfile targets for common operations:

```bash
# Apply all pending migrations
task migrate:up

# Roll back the last migration
task migrate:down

# Check current migration status
task migrate:status

# Create a new timestamped migration (SQL template)
task migrate:create name="add_table"
```

Each target delegates to the `migrate` subcommand defined in `internal/cli`, which opens a PostgreSQL connection using the configured `DB_URL` and executes the requested Goose action.

## API documentation

Swagger documentation can be (re)generated with:

```bash
task swagger:gen
```

This installs the `swag` CLI (if missing) and writes OpenAPI files into the `docs` directory, using `cmd/api/main.go` as the entry point.

## Testing and quality checks

Run the full test suite (with the race detector and coverage) via:

```bash
task test
```

For CI-parity without the race detector use `task test:ci`. Linting is provided through `task lint`, which runs `go vet` and `staticcheck` when available.

## Known issues

The repository currently contains several issues identified during initial auditing:

1. Trip service methods (`Get`, `Update`, `Buy`) return repository-level `ErrNotFound` errors directly. As the HTTP handlers look for `services.ErrTripNotFound`, missing trips result in `500 Internal Server Error` responses instead of `404`. Mapping repository errors to service errors before returning would fix the problem.
2. Authentication/profile services propagate `repository.ErrNotFound`, but the profile handler expects `services.ErrNotFound`, leading to the same 500 vs 404 mismatch for missing users.
3. The user repository never reads or writes the `avatar` column, so avatar updates disappear after saving and the value is absent in responses.

Addressing these items should be prioritized before shipping to production.

## Contributing

1. Fork the repository and create a feature branch.
2. Ensure code is formatted (`gofmt`) and passes `task lint` and `task test`.
3. Submit a pull request describing your changes and testing strategy.

## License

MIT (see repository owner for details).
