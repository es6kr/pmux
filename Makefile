.PHONY: build dev clean templ install setup fmt vet test

# Setup development environment
setup:
	git config core.hooksPath .githooks
	@echo "✓ Git hooks configured"

# Build binary
build: templ
	go build -o bin/pmux ./cmd/pmux

# Development with live reload
dev:
	@which air > /dev/null || go install github.com/air-verse/air@latest
	air

# Generate templ files
templ:
	@which templ > /dev/null || go install github.com/a-h/templ/cmd/templ@latest
	templ generate

# Clean build artifacts
clean:
	rm -rf bin/ tmp/

# Install binary to GOPATH/bin
install: build
	cp bin/pmux $(GOPATH)/bin/pmux

# Run tests
test:
	go test ./...

# Format code
fmt:
	gofmt -w .

# Run vet
vet:
	go vet ./...

# Download dependencies
deps:
	go mod download
	go mod tidy
