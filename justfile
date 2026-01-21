# hq - HUML Query Processor

default:
    @just --list

# Run all tests
test:
    go test ./...

# Run tests with verbose output
test-verbose:
    go test ./... -v

# Build hq binary
build:
    go build -o hq ./cmd/hq

# Install hq to GOPATH/bin
install:
    go install ./cmd/hq

# Clean build artifacts
clean:
    rm -f hq
    go clean

# Format code
fmt:
    go fmt ./...

# Run linter
lint:
    go vet ./...

# Generate walkthrough docs
docs:
    ./scripts/generate_walkthrough.sh

# Show test coverage
coverage:
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage.out
