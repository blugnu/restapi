package restapi

import "errors"

var (
	ErrBodyRequired            = errors.New("a body is required")
	ErrErrorReadingRequestBody = errors.New("error reading request body")
	ErrInvalidAcceptHeader     = errors.New("no formatter for content type")
	ErrInvalidArgument         = errors.New("invalid argument")
	ErrInvalidOperation        = errors.New("invalid operation")
	ErrInvalidStatusCode       = errors.New("invalid statuscode")
	ErrMarshalErrorFailed      = errors.New("error marshalling an Error response")
	ErrMarshalResultFailed     = errors.New("error marshalling response")
	ErrNoAcceptHeader          = errors.New("no Accept header")
	ErrUnexpectedField         = errors.New("unexpected field")
)
