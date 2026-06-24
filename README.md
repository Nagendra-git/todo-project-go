# Todo API (Go + MongoDB) — Standard Project Layout

A Todo CRUD API built the way a real Go service is structured in production: a thin `cmd/` entrypoint, layered `internal/` packages (config → db → repository → handlers → router), middleware, tests with a fake repository, and Docker support — following the community-standard [Go Project Layout](https://github.com/golang-standards/project-layout).

## Project structure

```
todo-go-app/
├── cmd/
│   └── server/
│       └── main.go              # Entrypoint: wires everything together, graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go            # Loads configs/application.properties, env overrides, validation
│   ├── db/
│   │   └── mongo.go             # MongoDB connection lifecycle (Connect/Disconnect)
│   ├── models/
│   │   └── todo.go              # Domain types (Todo, TodoUpdate)
│   ├── repository/
│   │   └── todo_repository.go   # TodoRepository interface + Mongo implementation
│   ├── handlers/
│   │   ├── todo_handler.go      # HTTP handlers, depend on repository interface
│   │   ├── todo_handler_test.go # Unit tests using a fake in-memory repository
│   │   └── response.go          # Shared JSON response helpers
│   ├── middleware/
│   │   ├── cors.go              # CORS middleware
│   │   └── logging.go           # Request logging middleware
│   └── router/
│       └── router.go            # Route → handler wiring
├── configs/
│   └── application.properties   # App configuration (Java-style key=value)
├── Dockerfile
├── docker-compose.yml            # Spins up API + MongoDB together
├── Makefile
├── go.mod
└── README.md
```

### Why this layout?

- **`cmd/`** holds only entrypoints — no business logic, just wiring.
- **`internal/`** can't be imported by other modules, which is how Go enforces "this is private to this app."
- **Repository pattern** — handlers depend on the `TodoRepository` interface, not MongoDB directly. This is what lets `todo_handler_test.go` test HTTP logic with a fake, in-memory repo instead of a real database.
- **Config as its own package** — typed, validated, env-override-aware, instead of scattered `os.Getenv` calls.
- **Middleware as composable functions** — CORS and logging wrap the router rather than being bolted into handlers.

## Running locally

### Option 1: Go directly (requires local MongoDB)

```bash
go mod tidy
make run
# or: go run ./cmd/server
```

### Option 2: Docker Compose (spins up Mongo for you)

```bash
make docker-up
# or: docker compose up --build
```

The API will be available at `http://localhost:8080`.

## Configuration

Edit `configs/application.properties`:

```properties
mongo.uri=mongodb://localhost:27017
mongo.database=tododb
mongo.collection=todos
mongo.timeout.seconds=10

server.port=8080
server.read_timeout.seconds=5
server.write_timeout.seconds=10

cors.allowed_origin=*
```

Any key can be overridden by an environment variable: uppercase the key and replace `.` with `_` (e.g. `mongo.uri` → `MONGO_URI`). This is how `docker-compose.yml` points the API at the `mongo` container without editing the properties file.

## API Endpoints

| Method | Endpoint       | Description       |
|--------|----------------|--------------------|
| GET    | /healthz       | Health check       |
| POST   | /todos         | Create a todo      |
| GET    | /todos         | List all todos     |
| GET    | /todos/{id}    | Get a single todo  |
| PUT    | /todos/{id}    | Update a todo      |
| DELETE | /todos/{id}    | Delete a todo      |

## Running tests

```bash
make test
# or: go test ./...
```

Handler tests run against a fake in-memory repository (`internal/handlers/todo_handler_test.go`), so they don't need a live MongoDB connection.

## Graceful shutdown

`cmd/server/main.go` listens for `SIGINT`/`SIGTERM` and shuts the HTTP server down cleanly, finishing in-flight requests before exiting — the same pattern you'd want before deploying this behind a container orchestrator.
