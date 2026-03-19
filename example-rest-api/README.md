# example-rest-api

A minimal Go REST API template that demonstrates best practices for building RESTful APIs, including structured logging, middleware patterns, graceful shutdown, API versioning, and Swagger documentation.

## Project structure

```
.
├── cmd/
│   └── main.go              # Application entry point
├── doc/
│   ├── doc.go               # Swagger package documentation
│   ├── api.go               # Swagger route definitions
│   └── swagger.json         # Generated OpenAPI specification
├── handlers/
│   ├── handlers.go          # API mux initialization
│   └── v1/
│       ├── v1.go            # v1 route registration and middleware setup
│       ├── v1_test.go       # Integration tests
│       └── hello/
│           ├── hello.go     # Hello endpoint handler
│           └── hello_test.go
├── middleware/
│   └── middleware.go        # Logging, compression, panic recovery
├── web/
│   ├── request.go           # Request utilities (path param extraction)
│   └── response.go          # JSON response helpers
├── Makefile
└── go.mod
```

## Prerequisites

- [Go](https://golang.org/dl/) 1.26+
- [Docker](https://www.docker.com/) (for Swagger generation and UI)

## Running

```bash
make run PORT=8080
```

Then test it:

```bash
$ curl http://localhost:8080/api/v1/hello

{"message":"Hello, World!"}
```

## API endpoints

| Method | Path             | Description         |
|--------|------------------|---------------------|
| GET    | `/api/v1/hello`  | Returns a greeting  |

## Middleware

The following middleware is applied to all routes:

- **Logger** - structured JSON logging of request method, path, remote address, and duration via `slog`
- **Compress** - automatic gzip compression for responses
- **PanicRecovery** - recovers from panics and logs stack traces

## Graceful shutdown

The server listens for `SIGINT` and `SIGTERM` signals, giving in-flight requests up to 5 seconds to complete before shutting down.

## Testing

```bash
make test
```

## Swagger documentation

Generate the Swagger specification:

```bash
make swagger
```

Launch the Swagger UI (available at http://localhost:80):

```bash
make swagger-ui
```

## Available Make targets

| Target       | Description                       |
|--------------|-----------------------------------|
| `help`       | Show all available targets        |
| `test`       | Run unit and integration tests    |
| `run`        | Run the API (`PORT` required)     |
| `swagger`    | Generate Swagger documentation    |
| `swagger-ui` | Launch Swagger UI via Docker      |
