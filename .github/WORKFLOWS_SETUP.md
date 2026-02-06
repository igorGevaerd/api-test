# GitHub Actions Pipelines Documentation

This document describes all GitHub Actions workflows configured for this project and how to set them up.

## Overview

The project uses four main CI/CD pipelines:

1. **CI/CD Pipeline** - Runs tests, linting, formatting checks, and builds
2. **Docker Pipeline** - Builds and pushes Docker images to GitHub Container Registry
3. **Security Pipeline** - Runs CodeQL, Gosec, Snyk, and Dependency Check
4. **Coverage Pipeline** - Generates test coverage reports and enforces coverage thresholds

## Workflows

### 1. CI/CD Pipeline (`.github/workflows/ci.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

**Jobs:**
- **Test**: Runs unit tests on Go 1.21 and 1.22, uploads coverage to Codecov
- **Lint**: Runs golangci-lint with timeout of 5 minutes
- **Build**: Compiles the binary (depends on test and lint passing)
- **Format Check**: Verifies code is properly formatted with gofmt
- **Vet**: Runs go vet for code analysis

**Status Checks:**
- All jobs must pass before merging to main/develop
- Tests run on both Go 1.21 and 1.22 for compatibility

**Environment Variables:** None required

### 2. Docker Pipeline (`.github/workflows/docker.yml`)

**Triggers:**
- Push to `main` branch
- Any tagged releases (v*)
- Pull requests to `main` branch (build only, no push)

**Jobs:**
- **Build**: Builds Docker image using multi-stage build from `docker/Dockerfile`
  - Pushes to GitHub Container Registry (ghcr.io)
  - Uses Docker layer caching for faster builds
  - Generates semantic version tags for releases
  
- **Scan**: Runs Trivy vulnerability scanner on the built image
  - Only runs on main branch pushes (not on PRs)
  - Uploads results to GitHub Security tab

**Registry:**
- Container Registry: `ghcr.io/YOUR_GITHUB_USERNAME/api-test`

**Permissions Required:**
- `contents: read`
- `packages: write`

**Notes:**
- Automatically authenticated via `GITHUB_TOKEN` (no secrets needed)
- Images tagged as:
  - `main-sha-xxx` for main branch
  - `v1.0.0` for version tags
  - `latest` for main branch

### 3. Security Pipeline (`.github/workflows/security.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches
- Weekly schedule (Sundays at 00:00 UTC)

**Jobs:**
- **CodeQL**: GitHub's native code analysis
  - Analyzes Go code for security vulnerabilities
  - Uploads results to Security tab

- **Gosec**: Go security checker
  - Scans for common security issues in Go code
  - Generates SARIF output

- **Dependency Check**: OWASP dependency vulnerability scanner
  - Checks dependencies for known vulnerabilities
  - Scans both direct and transitive dependencies

- **Snyk**: Snyk vulnerability and license scanning
  - Requires SNYK_TOKEN secret
  - Scans for high-severity vulnerabilities

**Required Secrets:**
- `SNYK_TOKEN` (optional, if using Snyk)

**Setup:**
1. For Snyk integration: Get token from https://app.snyk.io/account/settings
2. Add token to GitHub: Settings → Secrets and variables → Actions → New repository secret

### 4. Coverage Pipeline (`.github/workflows/coverage.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

**Jobs:**
- **Coverage**: Generates code coverage reports and enforces minimum threshold
  - Runs tests with coverage measurement
  - Uploads to Codecov
  - Comments on PRs with coverage changes
  - Enforces 70% coverage threshold
  - Generates HTML coverage report

**Coverage Threshold:** 70% (adjustable in workflow file)

**Artifacts:**
- `coverage-report` - HTML coverage report (retained for 30 days)

## Required Setup Steps

### 1. Enable GitHub Actions
1. Go to Settings → Actions → General
2. Ensure "Allow all actions and reusable workflows" is selected

### 2. Configure Branch Protection Rules
1. Go to Settings → Branches
2. Add protection rule for `main` branch:
   - ✅ Require status checks to pass before merging
   - ✅ Select required status checks:
     - `test (1.21)` and `test (1.22)`
     - `lint`
     - `build`
     - `fmt`
     - `vet`
   - ✅ Require PR reviews before merging
   - ✅ Dismiss stale PR approvals

### 3. Configure Codecov (Optional)
1. Go to https://codecov.io
2. Sign in with GitHub
3. Authorize Codecov to access your repository
4. Codecov will automatically receive coverage reports from workflows

### 4. Configure Snyk (Optional)
1. Go to https://app.snyk.io
2. Sign up/login with GitHub
3. Get your API token from Account Settings
4. Add to GitHub as secret:
   - Settings → Secrets and variables → Actions → New repository secret
   - Name: `SNYK_TOKEN`
   - Value: Your Snyk API token

### 5. Configure Container Registry
The Docker pipeline uses GitHub Container Registry (GHCR) which is automatically available:
- No additional setup needed
- Images will be pushed to `ghcr.io/YOUR_USERNAME/api-test`
- Set repository visibility as needed in Settings

## Common Tasks

### Viewing Workflow Results
1. Go to Actions tab
2. Select workflow from list
3. View summary and logs

### Debugging Failed Workflows
1. Click on failed workflow run
2. Expand the failed job step
3. Check logs for error messages
4. Common issues:
   - Tests failing: Check Go version compatibility
   - Coverage below threshold: Add more tests or adjust threshold
   - Docker build failing: Verify Dockerfile path and syntax

### Skipping Workflows
Add to commit message:
```
[skip ci]  # Skip all workflows
[skip docker]  # Skip Docker workflow (not supported by default)
```

### Running Locally
Test your changes before pushing:
```bash
# Run tests
make test

# Run linter
make lint

# Check formatting
gofmt -s -w .

# Run go vet
go vet ./...

# Build binary
make build

# Build Docker image
docker build -f docker/Dockerfile -t api-test:local .
```

## Secrets & Variables Reference

| Secret | Required | Used By | Source |
|--------|----------|---------|--------|
| `SNYK_TOKEN` | No | Security Pipeline | https://app.snyk.io/account/settings |
| `GITHUB_TOKEN` | Built-in | Docker & Security Pipelines | Automatically provided |

## Monitoring & Alerts

### Enable Email Notifications
1. Go to Settings → Notifications
2. Configure email alerts for:
   - Workflow run failures
   - Workflow run approvals

### Status Badge
Add to README.md:
```markdown
[![CI/CD](https://github.com/YOUR_USERNAME/api-test/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_USERNAME/api-test/actions/workflows/ci.yml)
[![Docker](https://github.com/YOUR_USERNAME/api-test/actions/workflows/docker.yml/badge.svg)](https://github.com/YOUR_USERNAME/api-test/actions/workflows/docker.yml)
[![Coverage](https://codecov.io/gh/YOUR_USERNAME/api-test/branch/main/graph/badge.svg)](https://codecov.io/gh/YOUR_USERNAME/api-test)
```

## Customization

### Changing Coverage Threshold
Edit `.github/workflows/coverage.yml`, line with `70`:
```yaml
if [ $COVERAGE -lt 70 ]; then  # Change 70 to desired percentage
```

### Adding New Test Matrix Versions
Edit `.github/workflows/ci.yml`:
```yaml
matrix:
  go-version: ['1.21', '1.22', '1.23']  # Add new versions
```

### Configuring Docker Registry
To use Docker Hub instead of GHCR, edit `.github/workflows/docker.yml`:
```yaml
registry: docker.io
username: ${{ secrets.DOCKERHUB_USERNAME }}
password: ${{ secrets.DOCKERHUB_TOKEN }}
```

## Troubleshooting

### Tests Failing Locally but Passing in CI
1. Ensure Go version matches (1.21)
2. Check environment variables (.env.example)
3. Run with same flags: `go test -v -race -timeout 10m ./...`

### Docker Build Fails in CI
1. Verify Dockerfile path: `docker/Dockerfile`
2. Check file permissions: `chmod +x docker/Dockerfile`
3. Review Dockerfile syntax

### Coverage Check Failing
1. Run locally: `go tool cover -func=coverage.out | grep total`
2. Add tests to increase coverage
3. Or adjust threshold in workflow file

### Codecov Not Receiving Reports
1. Verify Codecov is enabled: https://codecov.io/account/select/gh
2. Check workflow logs for upload errors
3. Manually trigger workflow in Actions tab

## Resources

- [GitHub Actions Documentation](https://docs.github.com/actions)
- [Go Testing](https://golang.org/pkg/testing/)
- [CodeQL](https://codeql.github.com/)
- [Codecov](https://codecov.io)
- [Snyk](https://snyk.io)
- [Trivy](https://github.com/aquasecurity/trivy)

## Next Steps

1. Push changes to repository
2. GitHub Actions will automatically start running workflows
3. Monitor Actions tab for results
4. Fix any failing status checks before merging

For questions or issues, refer to the workflow files in `.github/workflows/` directory.
