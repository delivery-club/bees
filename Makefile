lint:
	@echo "Running golangci-lint..."
	@golangci-lint run --config=.golangci.yml

test:
	@echo "Running tests..."
	@go test ./... -cover -short -count=1

ci-lint: install-linter lint

install-linter:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

