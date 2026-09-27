# MoolRoop (मूलरूप)

> **Minimalist Identity Core & Immutable Activity Engine**

MoolRoop is a high-performance, contract-first identity service built in Go. It enforces strict Personally Identifiable Information (PII) minimization and provides an append-only, immutable activity logging engine. The service concurrently exposes both a high-performance **gRPC** server and a **RESTful JSON HTTP gateway** with interactive **Swagger / OpenAPI** documentation via Gin.

---

## Key Features

- **Strict PII Minimization**: Eliminates unnecessary personal data (such as date of birth) and restricts contact data strictly to an email address.
- **Append-Only Immutable Activity Log**: Activity events are permanently recorded, time-stamped, and immutable. Reads return defensive copies to preserve integrity.
- **Contract-First Architecture**: Defined using Protocol Buffers (`proto3`) with automated Go stub and Swagger/OpenAPI generation.
- **Dual-Server Runtime**:
  - **gRPC Engine** on port `:50051` for high-throughput service-to-service communication.
  - **REST HTTP Engine + Swagger UI** on port `:8080` for JSON clients and interactive API testing.
- **Thread-Safe In-Memory Store**: Optimized using `sync.RWMutex` with race-tested safety across concurrent readers and writers.

---

## Project Structure

```
moolroop/
├── api/
│   └── proto/
│       └── v1/
│           └── identity.proto       # Service & message definitions (proto3)
├── cmd/
│   └── server/
│       └── main.go                  # Service bootstrapper (gRPC + Gin + Swagger)
├── docs/                            # Auto-generated Swagger documentation
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── errors/                      # Domain error definitions & gRPC status mappings
│   │   └── errors.go
│   ├── handler/                     # Gin REST & gRPC controller logic
│   │   ├── profile_handler.go       # REST profile controller
│   │   ├── activity_handler.go      # REST activity controller
│   │   ├── grpc_handler.go          # gRPC IdentityServiceServer implementation
│   │   └── handler_test.go          # REST & gRPC handler test suite
│   ├── model/                       # Domain models & validation logic
│   │   ├── profile.go               # Profile request payloads & validation
│   │   └── activity.go              # Activity request payloads & validation
│   └── store/                       # Thread-safe in-memory store
│       ├── memory.go                # RWMutex-backed store implementation
│       └── memory_test.go           # Concurrency and race tests
├── proto/gen/go/v1/                 # Generated protobuf structs and stubs
│   ├── identity.pb.go
│   └── identity_grpc.pb.go
├── gen/                             # Local package alias for proto stubs
├── buf.gen.yaml                     # Protobuf generator configuration
├── Makefile                         # Build, code-gen, test, and run routines
├── go.mod                           # Go module definition
├── go.sum                           # Checksums
├── LICENSE                          # MIT License
└── README.md                        # Project documentation
```

---

## Prerequisites

- **Go**: `1.22+` (tested with Go 1.24/1.27)
- **Protocol Buffer Compiler**: `protoc` (v36+)
- **Go Protobuf Plugins**:
  - `protoc-gen-go` (`go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`)
  - `protoc-gen-go-grpc` (`go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`)
- **Swag CLI**:
  - `swag` (`go install github.com/swaggo/swag/cmd/swag@latest`)

---

## Quickstart

### 1. Build and Run

You can build and start the dual-server using the `Makefile`:

```bash
# Compile proto stubs, generate swagger docs, run tests, and build binary
make all

# Start both gRPC and HTTP/REST servers
make run
```

Alternatively, run directly with Go:

```bash
go run cmd/server/main.go
```

The runtime will launch:
- **gRPC Server**: `localhost:50051`
- **HTTP / REST API**: `http://localhost:8080/api/v1`
- **Interactive Swagger UI**: `http://localhost:8080/swagger/index.html`

---

## API Reference

### REST Endpoints (`/api/v1`)

| Method | Endpoint | Description | Request Body / Params |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/profiles` | Create a new user profile | `{"full_name": "string", "email": "string"}` |
| `GET` | `/api/v1/profiles/:id` | Retrieve profile by user ID | Path parameter `:id` |
| `PATCH` | `/api/v1/profiles/:id` | Incremental profile update | `{"full_name"?: "string", "email"?: "string"}` |
| `POST` | `/api/v1/activities` | Append immutable activity log | `{"user_id": "string", "action_type": "string", "description"?: "string"}` |
| `GET` | `/api/v1/activities/:user_id` | List all activities for user | Path parameter `:user_id` |

### gRPC Service (`moolroop.v1.IdentityService`)

```protobuf
service IdentityService {
  rpc CreateProfile(CreateProfileRequest) returns (ProfileResponse);
  rpc GetProfile(GetProfileRequest) returns (ProfileResponse);
  rpc PatchProfile(PatchProfileRequest) returns (ProfileResponse);

  rpc LogActivity(LogActivityRequest) returns (ActivityResponse);
  rpc ListActivities(ListActivitiesRequest) returns (ListActivitiesResponse);
}
```

---

## Example Usage (cURL)

### 1. Create a Profile
```bash
curl -X POST http://localhost:8080/api/v1/profiles \
  -H "Content-Type: application/json" \
  -d '{"full_name": "Jane Doe", "email": "jane.doe@example.com"}'
```

**Response (`201 Created`):**
```json
{
  "user_id": "8c4df821-4d37-4dbe-a7db-27bbecb0bb67",
  "full_name": "Jane Doe",
  "email": "jane.doe@example.com",
  "created_at": { "seconds": 1790405896, "nanos": 507028000 },
  "updated_at": { "seconds": 1790405896, "nanos": 507028000 }
}
```

### 2. Patch a Profile
```bash
curl -X PATCH http://localhost:8080/api/v1/profiles/8c4df821-4d37-4dbe-a7db-27bbecb0bb67 \
  -H "Content-Type: application/json" \
  -d '{"full_name": "Jane Smith"}'
```

### 3. Log an Activity Record
```bash
curl -X POST http://localhost:8080/api/v1/activities \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "8c4df821-4d37-4dbe-a7db-27bbecb0bb67",
    "action_type": "USER_LOGIN",
    "description": "Authenticated via WebAuthn passkey"
  }'
```

### 4. List Activities
```bash
curl http://localhost:8080/api/v1/activities/8c4df821-4d37-4dbe-a7db-27bbecb0bb67
```

---

## Development & Makefile Targets

The included [`Makefile`](file:///Users/rohithtp/mine/home/workspaces/moolroop/Makefile) automates common development workflows:

```bash
make proto    # Recompile Protocol Buffer stubs
make swag     # Rebuild Swagger docs from handler annotations
make lint     # Run golangci-lint static analysis
make test     # Execute all test suites with race detection (-race)
make vet      # Run static analysis with go vet
make build    # Compile the binary to bin/server
make clean    # Remove build artifacts
```

---

## Testing

Run unit and concurrent race tests:

```bash
go test -v -race ./...
```

All store routines, profile handlers, and gRPC methods are covered with test suites verifying thread safety, validation, and error mappings.

---

## License

This project is licensed under the terms of the [MIT License](LICENSE).
