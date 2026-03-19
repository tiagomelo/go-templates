package doc

// swagger:route GET /api/v1/hello hello SayHello
// Say Hello.
// ---
// responses:
//		200: sayHelloResponse
//		500: description: internal server error

// swagger:response sayHelloResponse
type SayHelloResponseWrapper struct {
	// in:body
	Body SayHelloResponse
}

type SayHelloResponse struct {
	// example: Hello, World!
	Message string `json:"message"`
}
