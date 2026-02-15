VERSION ?= 1.0.0

# Build variables
LDFLAGS := -ldflags "-X github.com/prometheus/common/version.Version=$(VERSION)"

# Build the bitcoind_exporter
build:
	go build $(LDFLAGS) .

# Build with promu
build-promu:
	promu build --prefix .

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Run vet
vet:
	go vet ./...

# Format code
fmt:
	go fmt ./...

# Run all checks
check: fmt vet test

# Build all platforms
crossbuild:
	promu crossbuild

# Build Docker image
docker-build:
	docker build -t bitcoind_exporter:$(VERSION) .

# Push Docker image
docker-push:
	docker push bitcoind_exporter:$(VERSION)

.PHONY: build build-promu test test-coverage vet fmt check crossbuild docker-build docker-push
