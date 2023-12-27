# example-rest-api

A sample REST API project in Go designed for managing books. This project serves as a concrete example of a fully functional REST API without relying on complex frameworks.

## motivation

The motivation behind creating `example-rest-api` was to offer more than just a basic "hello world" example. I aimed to provide a fully functional REST API template, almost production-ready, to demonstrate real-world applications. This project addresses the gap often found in simplistic examples, offering a comprehensive and practical guide for developers looking to understand and implement a REST API in Go without overly complex frameworks.

## key features

- Uses [Gorilla Mux](https://github.com/gorilla/mux) for HTTP routing.
- Implements custom middleware and [Gorilla Handlers](https://github.com/gorilla/handlers).
- Input validation with [validator](https://github.com/go-playground/validator).
- Database migrations handled by [golang-migrate](https://github.com/golang-migrate/migrate).
- API documentation through [go-swagger](https://github.com/go-swagger/go-swagger).
- Ensures 100% test coverage, including both unit and integration tests.

## running it

```
make run PORT=<port>
```

## running tests

```
make test
```

## coverage report

```
make coverage
```

## api documentation

Two files hold api's documentation: [doc.go](doc/doc.go) and [api.go](doc/api.go).

To re-generate [doc/swagger.json](doc/swagger.json),

```
make swagger
```

To view it on a browser,

```
make swaggger-ui
```

then visit `localhost`.

## available `Makefile` targets

To generate the basic `Makefile` I've used [go-makefile-gen](https://github.com/tiagomelo/go-makefile-gen), a tool that I've written.