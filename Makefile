lint:
	@echo "Running golangci-lint..."
	@golangci-lint run --config=.golangci.yml

test:
	@echo "Running tests..."
	@go test ./... -cover -short -count=1

ci-lint: install-linter lint

install-linter:
	@go install  github.com/quasilyte/go-ruleguard/cmd/ruleguard@cb19258d2ade88dbf466420bb4585dc747bcec57
