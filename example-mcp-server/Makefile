.PHONY: help
## help: shows this help message
help:
	@ echo "Usage: make [target]\n"
	@ sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: test
## test: run unit tests
test:
	@ go test -v ./... -count=1

.PHONY: coverage
## coverage: run unit tests and generate coverage report (coverage.out + coverage.html)
coverage:
	@ set -e; \
	packages=$$(go list ./... | grep -v "/cmd" ); \
	if [ -z "$$packages" ]; then \
		echo "No valid Go packages found"; \
		exit 1; \
	fi; \
	echo "Packages:" $$packages; \
	go test -race -coverpkg=$$(echo $$packages | tr ' ' ',') -coverprofile=coverage.out $$packages; \
	go tool cover -html=coverage.out -o coverage.html; \
	echo "Generated: coverage.out coverage.html"

.PHONY: run
## run: run the MCP server
run:
	@ cat | go run cmd/main.go
