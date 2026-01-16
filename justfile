# Default recipe to display help
default:
    @just --list

# Build the binary
build:
    go build -o ccstats .

# Run tests
test:
    go test ./...

# Run tests with verbose output
test-v:
    go test -v ./...

# Run tests with coverage
test-cov:
    go test -cover ./...

# Format code
fmt:
    go fmt ./...

# Run the linter
lint:
    golangci-lint run

# Run the application
run *ARGS:
    go run . {{ARGS}}

# Install the binary to $GOPATH/bin
install:
    go install .

# Clean build artifacts
clean:
    rm -f ccstats

# Check everything (fmt, lint, test)
check: fmt lint test
