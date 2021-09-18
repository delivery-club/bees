lint:
	@echo "Running golangci-lint..."
	@golangci-lint run

test:
	@echo "Running tests..."
	@go test ./... -cover -short -count=1
