# Ceph Exporter Development Guide

## Pre-commit Setup

This project uses pre-commit hooks to ensure code quality.

### Installation

1. Install pre-commit (if not already installed):
```bash
pip install pre-commit
```

2. Install the git hooks:
```bash
pre-commit install
```

### Usage

Pre-commit will automatically run on `git commit`. To manually run all hooks:

```bash
pre-commit run --all-files
```

### Required Tools

Make sure you have these tools installed:

- `gofmt` (included with Go)
- `goimports`: `go install golang.org/x/tools/cmd/goimports@latest`
- `golangci-lint`: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

### Hooks Configured

- **General**: trailing whitespace, end-of-file fixer, YAML/JSON validation
- **Go**: gofmt, goimports, go vet, golangci-lint

## GitHub Actions

The project includes several CI workflows:

### 1. Pre-commit (`pre-commit.yml`)
Runs on every push and PR to main/develop branches.
Executes all pre-commit hooks to ensure code quality.

### 2. CI (`ci.yml`)
- **Build and Test**: Tests on Go 1.21 and 1.22
- **Lint**: Runs golangci-lint with project configuration
- **Coverage**: Uploads test coverage to Codecov

### 3. Integration Tests (`integration-test.yml`)
Runs integration tests with Docker Compose.

## Building the Project

```bash
cd ceph-exporter
go build -o ceph-exporter ./cmd/ceph-exporter
```

## Running Tests

```bash
cd ceph-exporter
go test -v ./...
```

## Code Quality

Run linters manually:

```bash
cd ceph-exporter
golangci-lint run --config ../.golangci.yml
```
