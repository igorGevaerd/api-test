# User Management API

A simple REST API built with Go's standard library for managing users. This project demonstrates basic CRUD operations and HTTP handling.

## Prerequisites

- Go 1.16 or higher
- curl or Postman (for testing API endpoints)

## Installation

1. Clone or navigate to the project directory:
```bash
cd api-test
```

2. Install dependencies:
```bash
make install-deps
```

## Running the Server

### Using Make (Recommended)

```bash
# Start with Docker Compose
make docker-up

# Run locally
make run

# Build executable
make build
```

### Using Go directly

```bash
go run ./cmd/api
```

The server will start on `http://localhost:8080`

You should see:
```
Server running on :8080
```

## Docker

### Building the Docker Image

Build the Docker image using the provided Dockerfile:

```bash
make docker-build
```

Or manually:
```bash
docker build -f docker/Dockerfile -t api-server .
```

### Running with Docker Compose (Recommended)

The easiest way to run the complete stack (API + PostgreSQL + Redis) is with Docker Compose:

#### 1. Create environment file (optional)

Copy the example environment file:
```bash
cp .env.example .env
```

Edit `.env` to customize database credentials if needed.

#### 2. Start all services

```bash
make docker-up
```

Or with docker-compose directly:
```bash
docker-compose -f docker/docker-compose.yml up -d
```

This will start:
- **API Server** on `http://localhost:8080`
- **PostgreSQL** on `localhost:5432`
- **Redis** on `localhost:6379`

#### 3. Verify services are running

```bash
docker-compose -f docker/docker-compose.yml ps
```

Or using make:
```bash
make docker-logs
```

#### 4. View logs

```bash
# All services
docker-compose -f docker/docker-compose.yml logs -f

# Specific service
docker-compose -f docker/docker-compose.yml logs -f api
docker-compose -f docker/docker-compose.yml logs -f db
docker-compose -f docker/docker-compose.yml logs -f redis
```

Or using make:
```bash
make docker-logs
make docker-logs-api
make docker-logs-db
make docker-logs-redis
```

#### 5. Stop all services

```bash
make docker-down
```

Or:
```bash
docker-compose -f docker/docker-compose.yml down
```

#### 6. Clean up volumes (remove data)

```bash
make docker-clean
```

Or:
```bash
docker-compose -f docker/docker-compose.yml down -v
```

### Running Individual Docker Container

To run only the API container (requires separate database and Redis):

```bash
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e REDIS_HOST=host.docker.internal \
  api-server
```

Or using make:
```bash
make docker-build
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e REDIS_HOST=host.docker.internal \
  api-server
```

### Docker Compose Services

#### PostgreSQL Database
- **Image:** postgres:15-alpine
- **Port:** 5432
- **Default User:** postgres
- **Default Password:** password
- **Default Database:** api_db
- **Volumes:** postgres_data (persistent storage)
- **Health Check:** Enabled

#### Redis Cache
- **Image:** redis:7-alpine
- **Port:** 6379
- **Configuration:** Persistence enabled (AOF)
- **Volumes:** redis_data (persistent storage)
- **Health Check:** Enabled

#### API Server
- **Built from:** Dockerfile
- **Port:** 8080
- **Environment:** Automatically configured from .env
- **Depends On:** Database and Redis (waits for health checks)
- **Network:** api-network (internal bridge network)

## API Endpoints

### 1. Get All Users
**Endpoint:** `GET /users`

Returns a list of all users.

**Example:**
```bash
curl http://localhost:8080/users
```

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com"
  },
  {
    "id": 2,
    "name": "Bob",
    "email": "bob@example.com"
  }
]
```

---

### 2. Get Single User
**Endpoint:** `GET /user?id={id}`

Returns a specific user by ID.

**Query Parameters:**
- `id` (required): The user ID to retrieve

**Example:**
```bash
curl http://localhost:8080/user?id=1
```

**Response (200 OK):**
```json
{
  "id": 1,
  "name": "Alice",
  "email": "alice@example.com"
}
```

**Response (404 Not Found):**
```json
{
  "error": "user not found"
}
```

---

### 3. Create User
**Endpoint:** `POST /users`

Creates a new user with the provided data.

**Request Body:**
```json
{
  "name": "Charlie",
  "email": "charlie@example.com"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'
```

**Response (201 Created):**
```json
{
  "email": "charlie@example.com"
}
```

---

## Testing

### Using Make

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific test
go test -v -run TestName ./...
```

### Using Docker Compose

Start the full stack first:
```bash
make docker-up
```

Wait a few seconds for all services to be healthy.

### Manual Testing with curl

#### Test 1: Health check
```bash
curl http://localhost:8080/health
```

Response:
```json
{"status":"ok"}
```

#### Test 2: Get all users (initially seeded from database)
```bash
curl http://localhost:8080/users
```

Response:
```json
[
  {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com",
    "created_at": "2024-02-06T10:30:00Z",
    "updated_at": "2024-02-06T10:30:00Z"
  },
  {
    "id": 2,
    "name": "Bob",
    "email": "bob@example.com",
    "created_at": "2024-02-06T10:30:00Z",
    "updated_at": "2024-02-06T10:30:00Z"
  }
]
```

Note: First call fetches from database, subsequent calls are served from Redis cache (5 minute TTL).

#### Test 3: Get a specific user
```bash
curl http://localhost:8080/user?id=1
```

Response:
```json
{
  "id": 1,
  "name": "Alice",
  "email": "alice@example.com",
  "created_at": "2024-02-06T10:30:00Z",
  "updated_at": "2024-02-06T10:30:00Z"
}
```

Note: First call fetches from database, subsequent calls are served from Redis cache (10 minute TTL).

#### Test 4: Create a new user
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'
```

Response (201 Created):
```json
{
  "id": 3,
  "name": "Charlie",
  "email": "charlie@example.com",
  "created_at": "2024-02-06T11:00:00Z",
  "updated_at": "2024-02-06T11:00:00Z"
}
```

Note: This invalidates the "all_users" cache, forcing refresh on next request.

#### Test 5: Create a new user (missing fields)
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"David"}'
```

Response (400 Bad Request):
```json
{"error":"name and email are required"}
```

#### Test 6: Get non-existent user
```bash
curl http://localhost:8080/user?id=999
```

Response (404 Not Found):
```json
{"error":"user not found"}
```

### Accessing Database Directly

Connect to PostgreSQL:
```bash
docker-compose -f docker/docker-compose.yml exec db psql -U postgres -d api_db

# Then query users
SELECT * FROM users;
```

### Accessing Redis Directly

Connect to Redis CLI:
```bash
docker-compose -f docker/docker-compose.yml exec redis redis-cli

# View cached data
KEYS *

# Get specific cache entry
GET "user:1"
GET "all_users"

# Delete cache entry
DEL "all_users"
```

### Performance Testing with Apache Bench

Install Apache Bench (if not already installed):
```bash
# macOS
brew install httpd

# Linux (Ubuntu)
sudo apt-get install apache2-utils
```

Test with concurrency:
```bash
# 1000 requests, 10 concurrent
ab -n 1000 -c 10 http://localhost:8080/users
```

The second run should be significantly faster due to Redis caching.

## Configuration

### Environment Variables

The API now uses environment variables for configuration. Default values are used if not set:

```
# Database Configuration
DB_HOST=localhost          # PostgreSQL host
DB_PORT=5432              # PostgreSQL port
DB_USER=postgres          # PostgreSQL username
DB_PASSWORD=password      # PostgreSQL password
DB_NAME=api_db            # PostgreSQL database name

# Cache Configuration
REDIS_HOST=localhost      # Redis host
REDIS_PORT=6379           # Redis port

# Server Configuration
PORT=8080                 # API server port
```

### Using .env File

For Docker Compose, create a `.env` file:

```bash
cp .env.example .env
```

Then customize as needed:

```bash
DB_USER=myuser
DB_PASSWORD=mypassword
DB_NAME=mydb
```

### Local Development (Without Docker)

Before running locally, ensure PostgreSQL and Redis are running. Then:

```bash
# Set environment variables (optional, defaults will be used)
export DB_HOST=localhost
export REDIS_HOST=localhost

# Download dependencies
go mod download

# Run the server
go run main.go
```

### Modifying Configuration

**To change the API port:**
```bash
export PORT=3000
make run
```

**To use a different database:**
```bash
export DB_HOST=db.example.com
export DB_USER=custom_user
export DB_PASSWORD=custom_pass
make run
```

## Troubleshooting

### Port Already in Use

**Error:** `listen tcp :8080: bind: address already in use`

**Solution:** Either free the port or change to a different one:
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use a different port
PORT=3000 make run
```

### Connection Refused

**Error:** `curl: (7) Failed to connect to localhost port 8080`

**Solution:** Ensure the server is running:
```bash
make run
```

### Invalid JSON Response

**Error:** Malformed JSON in request body

**Solution:** Ensure proper JSON formatting:
```bash
# ✅ Correct
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}'

# ❌ Incorrect (missing quotes)
curl -X POST http://localhost:8080/users \
  -d '{name:John,email:john@example.com}'
```

### User Not Found

**Error:** `{"error":"user not found"}` with status 404

**Solution:** Ensure the user ID exists. Get all users first:
```bash
curl http://localhost:8080/users
```

### Redis/Database Connection Issues

**Error:** `Failed to connect to Redis` or `Failed to connect to database`

**Solution:** Check if services are running:
```bash
make docker-logs
# Or check specific services
docker ps | grep api
```

## Development

### Building and Running

```bash
# Build executable
make build

# Run locally (requires DB and Redis)
make run

# Run tests
make test

# Generate coverage report
make test-coverage

# Format code
make fmt

# Run linter
make lint
```

### Full Development Setup

```bash
# Install deps, build, and start Docker services
make dev-setup

# View logs
make docker-logs

# Run tests
make test
```

### Code Style

- Format code: `make fmt`
- Run linter: `make lint`
- Follow Go conventions in [Effective Go](https://golang.org/doc/effective_go)

### Project Structure

```
api-test/
├── cmd/api/                    # Application entry point
│   └── main.go                 # Clean initialization
│
├── internal/                   # Private packages
│   ├── config/
│   │   └── config.go          # Configuration loading
│   ├── database/
│   │   └── db.go              # Database connection
│   ├── handler/
│   │   ├── user.go            # User HTTP handlers
│   │   ├── health.go          # Health check
│   │   └── routes.go          # Route registration
│   ├── model/
│   │   └── user.go            # Data structures
│   ├── service/
│   │   └── user_service.go    # Business logic
│   ├── cache/
│   │   └── redis.go           # Redis client wrapper
│   └── util/                  # Helper utilities
│
├── migrations/
│   └── 001_create_users_table.sql
│
├── docker/
│   ├── Dockerfile             # Multi-stage build
│   └── docker-compose.yml     # Services orchestration
│
├── tests/                     # Integration tests
├── docs/                      # API documentation
├── scripts/                   # Utility scripts
│
├── main_test.go               # Unit tests
├── Makefile                   # Build automation
├── CONTRIBUTING.md            # Contributing guidelines
├── .gitignore                 # Git ignore rules
├── .env.example               # Environment variables template
├── go.mod                     # Go module definition
├── README.md                  # This file
└── [other files]
```

### Adding New Features

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines on:
- Creating new endpoints
- Writing tests
- Database migrations
- Pull request process

## Code Structure

- `cmd/api/main.go` - Application entry point
- `internal/handler/` - HTTP handlers
- `internal/service/` - Business logic
- `internal/model/` - Data structures
- `internal/cache/` - Redis operations
- `internal/database/` - Database operations
- `internal/config/` - Configuration management
- `migrations/` - Database migrations
- `main_test.go` - Unit tests

## Making Changes

1. Add/edit files in appropriate `internal/` packages
2. Update handlers in `internal/handler/`
3. Add business logic in `internal/service/`
4. Write tests in relevant `*_test.go` files
5. Run tests: `make test`
6. Format code: `make fmt`

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for complete guidelines on:
1. Making changes to the code
2. Testing thoroughly
3. Following Go conventions
4. Submitting pull requests

## Data Storage

### Database
- **Type:** PostgreSQL 15
- **Schema:** `users` table with id, name, email, created_at, updated_at
- **Persistence:** Volume-based (postgres_data)
- **Initialization:** Automated via migrations on first run
- **Indexes:** Email field indexed for faster lookups

### Cache
- **Type:** Redis 7
- **Strategy:** 
  - All users: 5 minute TTL
  - Individual user: 10 minute TTL
- **Persistence:** AOF (Append Only File) enabled
- **Invalidation:** Cache cleared when new user is created

## Key Features

- ✅ PostgreSQL for persistent data storage
- ✅ Redis for caching and improved performance
- ✅ Docker and Docker Compose for easy deployment
- ✅ Health checks for all services
- ✅ Input validation and error handling
- ✅ Proper HTTP status codes
- ✅ Clean, modular code architecture
- ✅ Environment-based configuration
- ✅ Comprehensive test coverage
- ✅ Production-ready structure
- ✅ Build automation with Makefile

## Future Enhancements

- Implement DELETE and UPDATE operations
- Add authentication/authorization (JWT)
- Implement rate limiting
- Add structured logging (logrus/zap)
- Add metrics and monitoring (Prometheus)
- Use a framework like Gin or Echo for better routing
- Write integration tests
- Add database migrations management (migrate tool)
- Implement API versioning
- Add request/response validation schemas
- Implement pagination for list endpoints
- Add GraphQL support

## License

This project is provided as-is for educational purposes.

