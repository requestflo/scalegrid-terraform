// Package client implements a thin Go client for the ScaleGrid REST API
// (https://scalegrid.io/api/). It is intentionally self-contained so it can be
// reused outside of the Terraform provider and unit tested in isolation.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the public ScaleGrid API endpoint. It can be
	// overridden through Config.BaseURL (for example to target a dedicated or
	// staging environment).
	DefaultBaseURL = "https://api.scalegrid.io/v1"

	// DefaultTimeout is the per-request HTTP timeout.
	DefaultTimeout = 60 * time.Second

	defaultUserAgent = "terraform-provider-scalegrid"
)

// AuthMode selects how requests are authenticated against the ScaleGrid API.
type AuthMode string

const (
	// AuthBasic uses HTTP Basic authentication with the account email as the
	// username and the API key as the password. This is the default scheme
	// used by the ScaleGrid API.
	AuthBasic AuthMode = "basic"

	// AuthBearer sends the API key/token as a Bearer token in the
	// Authorization header.
	AuthBearer AuthMode = "bearer"
)

// Config holds everything required to build a Client.
type Config struct {
	// BaseURL is the API root. Defaults to DefaultBaseURL when empty.
	BaseURL string

	// Email is the ScaleGrid account email. Required for AuthBasic.
	Email string

	// APIKey is the secret API key generated from the ScaleGrid console.
	APIKey string

	// AuthMode selects the authentication scheme. Defaults to AuthBasic.
	AuthMode AuthMode

	// UserAgent is appended to the default User-Agent string.
	UserAgent string

	// HTTPClient lets callers inject a custom *http.Client (timeouts,
	// transports, test servers). Defaults to a client with DefaultTimeout.
	HTTPClient *http.Client
}

// Client talks to the ScaleGrid REST API.
type Client struct {
	baseURL    string
	email      string
	apiKey     string
	authMode   AuthMode
	userAgent  string
	httpClient *http.Client
}

// NewClient validates the supplied configuration and returns a ready Client.
func NewClient(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("scalegrid: an API key is required")
	}

	mode := cfg.AuthMode
	if mode == "" {
		mode = AuthBasic
	}
	if mode == AuthBasic && cfg.Email == "" {
		return nil, errors.New("scalegrid: an account email is required for basic authentication")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultTimeout}
	}

	ua := defaultUserAgent
	if cfg.UserAgent != "" {
		ua = fmt.Sprintf("%s %s", cfg.UserAgent, defaultUserAgent)
	}

	return &Client{
		baseURL:    baseURL,
		email:      cfg.Email,
		apiKey:     cfg.APIKey,
		authMode:   mode,
		userAgent:  ua,
		httpClient: httpClient,
	}, nil
}

// APIError represents a non-2xx response from the ScaleGrid API.
type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	Errors     string `json:"errors,omitempty"`
	RawBody    string `json:"-"`
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = e.Errors
	}
	if msg == "" {
		msg = e.RawBody
	}
	return fmt.Sprintf("scalegrid api error (status %d): %s", e.StatusCode, msg)
}

// IsNotFound reports whether err is an APIError with a 404 status code. It is
// used by the provider to remove resources from state when they vanish.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// do performs an HTTP request against path (relative to the base URL),
// JSON-encoding body when non-nil and JSON-decoding the response into out when
// non-nil. It returns an *APIError for non-2xx responses.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("scalegrid: encoding request body: %w", err)
		}
		reqBody = bytes.NewReader(encoded)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("scalegrid: building request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.authenticate(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("scalegrid: performing %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("scalegrid: reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, respBytes)
	}

	if out != nil && len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, out); err != nil {
			return fmt.Errorf("scalegrid: decoding response: %w", err)
		}
	}
	return nil
}

func (c *Client) authenticate(req *http.Request) {
	switch c.authMode {
	case AuthBearer:
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	default: // AuthBasic
		req.SetBasicAuth(c.email, c.apiKey)
	}
}

func parseAPIError(status int, body []byte) error {
	apiErr := &APIError{StatusCode: status, RawBody: string(body)}
	// Best-effort decode of a structured error payload; ignore failures and
	// fall back to the raw body.
	_ = json.Unmarshal(body, apiErr)
	return apiErr
}
