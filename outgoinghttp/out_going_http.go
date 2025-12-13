package outgoinghttp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/goccy/go-json"
)

type RequestBuilder func(ctx context.Context) (*http.Request, error)

func newHTTPRequest[T any](
	ctx context.Context,
	httpMethod string,
	url string,
	body T,
) (*http.Request, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(b)

	req, err := http.NewRequestWithContext(ctx, httpMethod, url, bodyReader)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func BuildBasicRequest(
	httpMethod string,
	url string,
	opts ...AdditionalHTTPOptions,
) RequestBuilder {
	return func(ctx context.Context) (*http.Request, error) {
		args := bindAdditionalHTTPOptions(opts...)
		req, err := newHTTPRequest(ctx, httpMethod, url, args.Body)
		if err != nil {
			return nil, err
		}

		req.URL.RawQuery = args.Query.Encode()
		req.Header = args.Headers
		req.Header.Add("Content-Type", "application/json")

		return req, nil
	}
}

// CallHTTP satisfied the HTTPService interface by implementing basic calls
func CallHTTP[T any](
	ctx context.Context,
	client *http.Client,
	reqFn RequestBuilder,
) (T, int, error) {
	var result T

	req, err := reqFn(ctx)
	if err != nil {
		return result, 0, err
	}

	response, err := client.Do(req)
	if response != nil {
		defer func() {
			err = response.Body.Close()
			if err != nil {
				slog.Warn("unable to close response.Body", err)
			}
		}()
	}
	if err != nil || response == nil {
		return result, 0, err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		if response.Body == nil {
			return result, response.StatusCode, nil
		}
		byteArr, err := io.ReadAll(response.Body)
		if err != nil {
			return result, response.StatusCode, err
		}

		return result, response.StatusCode, fmt.Errorf(
			"code: %d err: %w",
			response.StatusCode,
			errors.New(string(byteArr)),
		)
	}

	if response.Body == http.NoBody {
		return result, response.StatusCode, nil
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return result, response.StatusCode, err
	}

	err = json.Unmarshal(responseBytes, &result)
	return result, response.StatusCode, err
}
