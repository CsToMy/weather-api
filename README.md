# weather-api

A small REST API written in Go that returns the current weather for a city. It wraps the [OpenWeather](https://openweathermap.org/api) current weather endpoint and exposes a simpler, stable response of its own.

The project is also a practice ground for building a production-style Go service: clean package layout, dependency injection through interfaces, test-first development, linting, and timeouts on every network boundary.

## Features

- `GET /weather?city=<name>` returns temperature and a short description
- `GET /health` for liveness checks
- Clear HTTP status codes for client errors, unknown cities and upstream failures
- Timeouts on both the outgoing HTTP client and the server
- The API key is read from the environment and never appears in error messages or logs
- Standard library only, no web framework

## Requirements

- Go 1.27 or newer
- An OpenWeather API key (free tier is enough). A newly created key can take a while to activate and may return `401` until then.
- Optional: `make` and [`golangci-lint`](https://golangci-lint.run/) for the development shortcuts below

## Quick start

Set the API key and start the server.

PowerShell:

```powershell
$env:OPENWEATHER_API_KEY = "your-key"
go run ./cmd/server
```

Bash:

```bash
export OPENWEATHER_API_KEY="your-key"
go run ./cmd/server
```

The server listens on port `8080`. Try it from another terminal:

```bash
curl "localhost:8080/weather?city=Berlin"
```

On Windows PowerShell use `curl.exe` instead of `curl`.

## Run with Docker

```bash
cp .env.example .env      # then put your API key into .env
docker compose up --build
```

The API is then available at `http://localhost:8080`.

## API

### `GET /weather?city=<name>`

Example response (`200 OK`):

```json
{
  "city": "Berlin",
  "temp_c": 12.5,
  "description": "overcast clouds"
}
```

| Status | When | Body |
|--------|------|------|
| `200` | Weather found | weather object (see above) |
| `400` | `city` parameter missing | `{"error":"city parameter is required"}` |
| `404` | City not found | `{"error":"city not found"}` |
| `502` | OpenWeather unreachable, slow, or returned an unexpected response | `{"error":"weather service unavailable"}` |

Upstream details are only written to the server log, never returned to the client.

### `GET /health`

Returns `200` with the body `ok`.

## Configuration

| Variable | Required | Description |
|----------|----------|-------------|
| `OPENWEATHER_API_KEY` | yes | Your OpenWeather API key. The server refuses to start without it. |

Do not commit the key. `.env` files are already listed in `.gitignore`.

## Development

Common tasks are available through `make`:

| Command | Description |
|---------|-------------|
| `make run` | Start the server |
| `make build` | Build the binary into `bin/` |
| `make test` | Run all tests |
| `make cover` | Run all tests with coverage |
| `make vet` | Run `go vet` |
| `make fmt` | Format the code |
| `make lint` | Run `golangci-lint` |
| `make check` | Run `vet`, `lint` and `test` (use before committing) |

Without `make`, the equivalents are `go test ./...`, `go vet ./...` and `golangci-lint run`.

### Tests

The tests need no internet connection and no API key:

- The HTTP handler is tested against a fake weather provider, covering success, missing parameter, unknown city and upstream failure.
- The OpenWeather client is tested against a local `httptest` server, covering the success path, a `404`, a `500`, invalid JSON, request timeout via `context`, and that the API key does not leak into error messages.

Both use table-driven tests where it makes sense.

## Project structure

```
.
├── cmd/
│   └── server/          # program entry point (wiring and configuration)
├── internal/
│   ├── api/             # HTTP router and handlers
│   ├── openweather/     # OpenWeather client
│   └── weather/         # domain types and the Provider interface
├── .golangci.yml        # linter configuration
├── Makefile
└── go.mod
```

## Design notes

- **Interface at the consumer.** The handler depends on the small `weather.Provider` interface, not on OpenWeather. The real client and the test fakes are interchangeable, which keeps the handler tests fast and deterministic.
- **Dependencies are injected.** `main` creates the HTTP client, the OpenWeather client and the router and connects them. Nothing is a global.
- **`context` everywhere.** Requests carry the caller's `context`, so cancellations and timeouts propagate to the outgoing call.
- **Errors are mapped deliberately.** A sentinel error (`weather.ErrCityNotFound`) separates "no such city" (`404`) from every other failure (`502`). Wrapped errors keep the cause available through `errors.Is`.
- **Secrets stay out of errors.** The OpenWeather API accepts the key only as a URL parameter, and Go's HTTP client errors include the full URL. The client therefore unwraps the error before returning it.

## Roadmap

- [x] GitHub Actions: lint, test (with the race detector) and build on every push
- [x] Dockerfile and `docker-compose`
- [ ] PostgreSQL: saved cities (CRUD) and request history
- [ ] Response cache (in-memory first, Redis later)
- [ ] Graceful shutdown
- [ ] Small CLI in `cmd/` reusing the same packages