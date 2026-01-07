# Makefile for go-opcua-sim

.PHONY: all build server client clean run-server run-client test

# Build all binaries
all: build

# Build both server and client
build: server client

# Build server
server:
	@echo "Building OPC UA server..."
	@mkdir -p bin
	@go build -o bin/server ./cmd/server
	@echo "Server built successfully: bin/server"

# Build client
client:
	@echo "Building OPC UA client..."
	@mkdir -p bin
	@go build -o bin/client ./cmd/client
	@echo "Client built successfully: bin/client"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "Clean complete"

# Run server with default settings
run-server:
	@echo "Running OPC UA server..."
	@./bin/server -config sensors.json -script plc_logic.lua -scantime 100 -endpoint "opc.tcp://0.0.0.0:4840"

# Run server without PLC logic
run-server-no-plc:
	@echo "Running OPC UA server without PLC logic..."
	@./bin/server -config sensors.json -plc=false -endpoint "opc.tcp://0.0.0.0:4840"

# Run client to read a specific node
run-client:
	@echo "Running OPC UA client..."
	@./bin/client -endpoint "opc.tcp://localhost:4840" -node "ns=2;i=1000"

# Run client in continuous mode
run-client-continuous:
	@echo "Running OPC UA client in continuous mode..."
	@./bin/client -endpoint "opc.tcp://localhost:4840" -node "ns=2;i=1000" -continuous -interval 1000

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Dependencies downloaded"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "Dependencies tidied"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "Tests complete"

# Display help
help:
	@echo "Available targets:"
	@echo "  all              - Build all binaries (default)"
	@echo "  build            - Build both server and client"
	@echo "  server           - Build OPC UA server only"
	@echo "  client           - Build OPC UA client only"
	@echo "  clean            - Remove build artifacts"
	@echo "  run-server       - Run OPC UA server with default settings"
	@echo "  run-server-no-plc - Run OPC UA server without PLC logic"
	@echo "  run-client       - Run OPC UA client (single read)"
	@echo "  run-client-continuous - Run OPC UA client (continuous read)"
	@echo "  deps             - Download dependencies"
	@echo "  tidy             - Tidy dependencies"
	@echo "  fmt              - Format code"
	@echo "  test             - Run tests"
	@echo "  help             - Display this help message"
