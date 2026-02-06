# Contributing to API Test

Thank you for your interest in contributing to this project!

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/yourusername/api-test.git
   cd api-test
   ```

3. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

4. **Set up development environment**:
   ```bash
   make dev-setup
   ```

## Development Workflow

### Code Style

- Follow Go conventions as defined in [Effective Go](https://golang.org/doc/effective_go)
- Use `make fmt` to format your code:
  ```bash
  make fmt
  ```
- Run linter before committing:
  ```bash
  make lint
  ```

### Testing

All new features must include tests. Run tests before submitting:

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific test
go test -v -run TestName ./...
```

Aim for at least 80% code coverage on new code.

### Commit Messages

Write clear, descriptive commit messages:

```
Short summary (50 chars or less)

Longer description if needed, explaining:
- What changed
- Why it changed
- Any side effects
```

Example:
```
Add user update endpoint

Implement PUT /users/:id endpoint to allow updating user details.
Also adds UpdatedAt timestamp handling.
```

## Project Structure

```
api-test/
├── cmd/api/              # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database operations
│   ├── handler/         # HTTP handlers
│   ├── model/           # Data structures
│   ├── service/         # Business logic
│   ├── cache/           # Caching logic
│   └── util/            # Utilities
├── migrations/          # Database migrations
├── tests/               # Integration tests
└── Makefile            # Build commands
```

## Adding New Features

### Adding a New Endpoint

1. **Create handler** in `internal/handler/`:
   ```go
   // internal/handler/post.go
   func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
       // Implementation
   }
   ```

2. **Add service logic** in `internal/service/`:
   ```go
   // internal/service/post_service.go
   func (s *PostService) Create(ctx context.Context, post *model.Post) error {
       // Implementation
   }
   ```

3. **Create model** in `internal/model/`:
   ```go
   // internal/model/post.go
   type Post struct {
       ID    int
       Title string
       // ...
   }
   ```

4. **Register route** in `internal/handler/routes.go`:
   ```go
   http.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
       if r.Method == http.MethodPost {
           postHandler.Create(w, r)
       }
   })
   ```

5. **Write tests** in `main_test.go` or `tests/`:
   ```go
   func TestCreatePost(t *testing.T) {
       // Test implementation
   }
   ```

### Adding a Database Migration

1. Create a new SQL file in `migrations/`:
   ```bash
   migrations/002_add_posts_table.sql
   ```

2. Write your SQL:
   ```sql
   CREATE TABLE posts (
       id SERIAL PRIMARY KEY,
       user_id INTEGER NOT NULL REFERENCES users(id),
       title VARCHAR(255) NOT NULL,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );
   ```

## Before Submitting a PR

1. **Update tests**: Ensure all tests pass
   ```bash
   make test
   ```

2. **Check coverage**: Maintain or improve code coverage
   ```bash
   make test-coverage
   ```

3. **Format code**: Auto-format before committing
   ```bash
   make fmt
   ```

4. **Run linter**: Fix any linting issues
   ```bash
   make lint
   ```

5. **Test locally**: Run the full application
   ```bash
   make dev-setup
   curl http://localhost:8080/health
   ```

6. **Update documentation**: If you changed behavior, update README.md

## Pull Request Guidelines

1. **Clear title**: Describe what you changed
   - ✅ "Add DELETE endpoint for users"
   - ❌ "Fix stuff"

2. **Detailed description**:
   - What problem does this solve?
   - What changes were made?
   - Any breaking changes?
   - Testing instructions

3. **Link related issues**:
   ```
   Fixes #123
   ```

4. **Keep it focused**: One feature per PR

5. **Update CHANGELOG**: Add an entry under "Unreleased"

## Reporting Issues

When reporting bugs, include:

1. **Description**: Clear explanation of the issue
2. **Steps to reproduce**: Exact steps to trigger the bug
3. **Expected behavior**: What should happen
4. **Actual behavior**: What actually happens
5. **Environment**: 
   - Go version (`go version`)
   - OS and version
   - Docker version (if applicable)
6. **Logs**: Relevant error messages or logs

## Questions?

- Open a GitHub discussion
- Check existing issues first
- Ask in PRs or issues

## Code of Conduct

- Be respectful
- Welcome diverse perspectives
- Give and receive feedback gracefully
- Focus on the code, not the person

Thank you for contributing! 🎉
