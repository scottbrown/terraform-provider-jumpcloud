// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const defaultBaseURL = "https://console.jumpcloud.com/api/v2"
const defaultBaseURLV1 = "https://console.jumpcloud.com/api"

type Client struct {
	apiKey     string
	orgID      string
	baseURL    string
	baseURLV1  string
	httpClient *http.Client
}

func New(apiKey, orgID, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	baseURLV1 := defaultBaseURLV1
	if baseURL != defaultBaseURL {
		baseURLV1 = baseURL
	}
	return &Client{
		apiKey:     apiKey,
		orgID:      orgID,
		baseURL:    baseURL,
		baseURLV1:  baseURLV1,
		httpClient: &http.Client{},
	}
}

func (c *Client) doV1(ctx context.Context, method, path string, body any, result any) error {
	return c.doWithBase(ctx, c.baseURLV1, method, path, body, result)
}

func (c *Client) do(ctx context.Context, method, path string, body any, result any) error {
	return c.doWithBase(ctx, c.baseURL, method, path, body, result)
}

func (c *Client) doPatch(ctx context.Context, path string, body any, result any) error {
	return c.doWithBase(ctx, c.baseURL, "PATCH", path, body, result)
}

func (c *Client) doWithBase(ctx context.Context, baseURL, method, path string, body any, result any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.orgID != "" {
		req.Header.Set("x-org-id", c.orgID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	return nil
}

type PaginationParams struct {
	Limit int
	Skip  int
}

func (c *Client) buildPaginatedPath(basePath string, params *PaginationParams) string {
	if params == nil {
		return basePath
	}
	v := url.Values{}
	if params.Limit > 0 {
		v.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Skip > 0 {
		v.Set("skip", strconv.Itoa(params.Skip))
	}
	if len(v) == 0 {
		return basePath
	}
	return basePath + "?" + v.Encode()
}

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("JumpCloud API error (status %d): %s", e.StatusCode, e.Body)
}

func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}
