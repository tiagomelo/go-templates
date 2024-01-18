# ==============================================================================
# Help

.PHONY: help
## help: shows this help message
help:
	@ echo "Usage: make [target]\n"
	@ sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# ==============================================================================
# Database migrations

.PHONY: migrate-setup
## migrate-setup: installs golang-migrate
migrate-setup:
	@ go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: create-migrations
## create-migration: creates up and down migration files for a given name (make create-migrations NAME=<desired_name>)
create-migration: migrate-setup
	@ if [ -z "$(NAME)" ]; then echo >&2 please set the name of the migration via the variable NAME; exit 2; fi
	@ migrate create -ext sql -dir db/migrations -seq -digits 4 $(NAME)

.PHONY: migrate-up
## migrate-up: runs up N migrations, N is optional (make migrate-up N=<desired_migration_number>)
migrate-up: migrate-setup
	@ migrate -database 'sqlite3://db/booksGrpcService.db?query' -path db/migrations up $(N)

.PHONY: migrate-down
## migrate-down: runs down N migrations, N is optional (make migrate-down N=<desired_migration_number>)
migrate-down: migrate-setup
	@ migrate -database 'sqlite3://db/booksGrpcService.db?query' -path db/migrations down $(N)

.PHONY: migrate-to-version
## migrate-to-version: migrates to version V (make migrate-to-version V=<desired_version>)
migrate-to-version: migrate-setup
	@ if [ -z "$(V)" ]; then echo >&2 please set the desired version via the variable V; exit 2; fi
	@ migrate -database 'sqlite3://db/booksGrpcService.db?query' -path db/migrations goto $(V)

.PHONY: migrate-force-version
## migrate-force-version: forces version V (make migrate-force-version V=<desired_version>)
migrate-force-version: migrate-setup
	@ if [ -z "$(V)" ]; then echo >&2 please set the desired version via the variable V; exit 2; fi
	@ migrate -database 'sqlite3://db/booksGrpcService.db?query' -path db/migrations force $(V)

.PHONY: migrate-version
## migrate-version: checks current database migrations version
migrate-version: migrate-setup
	@ migrate -database 'sqlite3://db/booksGrpcService.db?query' -path db/migrations version

.PHONY: migrate-test-up
## migrate-test-up: runs up N migrations on test db, N is optional (make migrate-up N=<desired_migration_number>)
migrate-test-up: migrate-setup
	@ migrate -database 'sqlite3://db/booksRestApiTest.db?query' -path db/migrations up $(N)

.PHONY: migrate-test-down
## migrate-test-down: runs down N migrations on test db, N is optional (make migrate-down N=<desired_migration_number>)
migrate-test-down: migrate-setup
	@ migrate -database 'sqlite3://db/booksRestApiTest.db?query' -path db/migrations down $(N)

# ==============================================================================
# Proto

.PHONY: proto
## proto: compiles .proto files
proto:
	@ docker run -v $(PWD):/defs namely/protoc-all -i api/proto -f book.proto -l go -o api/proto/gen/book --go-source-relative

# ==============================================================================
# Tests

.PHONY: test
## test: run unit tests
test:
	@ go test -v ./... -count=1

.PHONY: coverage
## coverage: run unit tests and generate coverage report in html format
coverage:
	@ go test -coverprofile=coverage.out ./...  && go tool cover -html=coverage.out

# ==============================================================================
# App's execution

.PHONY: run
## run: runs the gRPC server
run: migrate-up
	@ go run cmd/main.go -p $(PORT)

# ==============================================================================
# TLS

.PHONY: gen-certs
## gen-certs: generate both certificate and private key files
gen-certs:
	@ rm -f cert/*.pem cert/*.srl ; \
	./cert/gen.sh