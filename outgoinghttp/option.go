package outgoinghttp

import (
	"net/http"
	"net/url"
)

type AdditionalHTTPArgs struct {
	Headers               http.Header
	Query                 url.Values
	Body                  any
	RequiredFieldsInBody  []ValidateRequireFields
	RequiredFieldsInQuery []ValidateRequireFields
}

type AdditionalHTTPOptions func(args *AdditionalHTTPArgs)

type ValidateRequireFields struct {
	Condition      func() bool
	RequiredFields []string
}

func WithAdditionalHeaders(header http.Header) AdditionalHTTPOptions {
	return func(args *AdditionalHTTPArgs) {
		args.Headers = header
	}
}

func WithAdditionalQuery(query url.Values) AdditionalHTTPOptions {
	return func(args *AdditionalHTTPArgs) {
		args.Query = query
	}
}

func WithAdditionalBody(body any) AdditionalHTTPOptions {
	return func(args *AdditionalHTTPArgs) {
		args.Body = body
	}
}

func bindAdditionalHTTPOptions(opts ...AdditionalHTTPOptions) AdditionalHTTPArgs {
	args := &AdditionalHTTPArgs{
		Headers: make(http.Header),
		Query:   make(url.Values),
	}
	for _, opt := range opts {
		opt(args)
	}
	return *args
}
