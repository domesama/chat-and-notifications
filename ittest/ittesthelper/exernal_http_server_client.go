package ittesthelper

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TestHTTPClient struct {
	client  *http.Client
	t       *testing.T
	baseURL string
}

func NewTestHTTPClient(t *testing.T, baseURL string) *TestHTTPClient {
	return &TestHTTPClient{
		client:  http.DefaultClient,
		t:       t,
		baseURL: baseURL,
	}
}

func (c *TestHTTPClient) CreateRequest(ctx context.Context, method string, path string, body string) *http.Request {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, strings.NewReader(body))
	require.NoError(c.t, err)

	return req
}

func (c *TestHTTPClient) CreateJSONRequest(ctx context.Context, method string, path string, body any) *http.Request {
	bodyStr, ok := body.(string)
	if !ok {
		bytes, err := json.Marshal(body)
		require.NoError(c.t, err)

		bodyStr = string(bytes)
	}

	req := c.CreateRequest(ctx, method, path, bodyStr)
	req.Header.Add("content-type", "application/json")

	return req
}

// HealthCheck performs health check on the given request until it gets 200 OK or times out
func (c *TestHTTPClient) HealthCheck(req *http.Request) {
	for i := 0; i < 20; i++ {
		resp, err := c.client.Do(req)
		if err != nil {
			continue
		}

		_, err = io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		err = resp.Body.Close()
		if err != nil {
			continue
		}

		if resp.StatusCode == 200 {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}
}
