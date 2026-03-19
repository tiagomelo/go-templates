# ==============================================================================
# Help

.PHONY: help
## help: shows this help message
help:
	@ echo "Usage: make [target]\n"
	@ sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# ==============================================================================
# Tests

.PHONY: test
## test: run unit tests
test: 
	@ go test -v ./... -count=1

# ==============================================================================
# Swagger

.PHONY: swagger
## swagger: generates api's documentation
swagger: 
	@ unset `env|grep DOCKER|cut -d\= -f1` ;\
	docker run --rm --name books-swagger -it -v $(HOME):$(HOME) -w $(PWD) quay.io/goswagger/swagger generate spec -o doc/swagger.json

.PHONY: swagger-ui
## swagger-ui: launches swagger ui
swagger-ui: swagger
	@ docker run --rm --name books-swagger-ui -p 80:8080 -e SWAGGER_JSON=/docs/swagger.json -v $(shell pwd)/doc:/docs swaggerapi/swagger-ui

# ==============================================================================
# App's execution

.PHONY: run
## run: runs the API
run:
	@ if [ -z "$(PORT)" ]; then echo >&2 please set the desired port via the variable PORT; exit 2; fi
	@ go run cmd/main.go -p $(PORT)