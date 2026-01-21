package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Retry settings
const (
	maxRetries   = 20
	baseWait     = 100 * time.Millisecond
	maxWait      = 5 * time.Second
	minWait      = 50 * time.Millisecond
	adjustFactor = 1.5
)

// VtexClient handles communication with the VTEX API
type VtexClient struct {
	vtexBaseURL   string
	oktaURL       string
	oktaClientID  string
	oktaSecret    string
	oktaGrantType string
	oktaScope     string
	httpClient    *http.Client
	token         string
	tokenExpiry   time.Time
	tokenMutex    sync.RWMutex
}

// UserRole represents a user with a role in VTEX
type UserRole struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Account  string `json:"account"`
	RoleName string `json:"roleName"`
}

// UserRoleRequest is the payload to create or delete users
type UserRoleRequest struct {
	Users []UserRole `json:"users"`
}

// OktaTokenResponse is the token response from Okta
type OktaTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// NewVtexClient creates a new VTEX client
func NewVtexClient(vtexBaseURL, oktaURL, oktaClientID, oktaSecret, oktaGrantType, oktaScope string) (*VtexClient, error) {
	return &VtexClient{
		vtexBaseURL:   vtexBaseURL,
		oktaURL:       oktaURL,
		oktaClientID:  oktaClientID,
		oktaSecret:    oktaSecret,
		oktaGrantType: oktaGrantType,
		oktaScope:     oktaScope,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// getToken gets a valid token, renews it if needed
func (c *VtexClient) getToken() (string, error) {
	c.tokenMutex.RLock()
	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		token := c.token
		c.tokenMutex.RUnlock()
		return token, nil
	}
	c.tokenMutex.RUnlock()

	// Need to renew the token
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	// Check again in case another goroutine already renewed it
	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return c.token, nil
	}

	// Get new token
	data := url.Values{}
	data.Set("grant_type", c.oktaGrantType)
	data.Set("scope", c.oktaScope)

	req, err := http.NewRequest("POST", c.oktaURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.oktaClientID, c.oktaSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error requesting token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error obtaining token: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp OktaTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("error decoding token response: %w", err)
	}

	c.token = tokenResp.AccessToken
	// Set expiry with 5 minutes margin
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	return c.token, nil
}

// refreshToken forces token renewal
func (c *VtexClient) refreshToken() (string, error) {
	c.tokenMutex.Lock()
	c.token = ""
	c.tokenExpiry = time.Time{}
	c.tokenMutex.Unlock()
	return c.getToken()
}

// doRequestWithRetry runs a request with retries and exponential backoff
func (c *VtexClient) doRequestWithRetry(method, endpoint string, payload interface{}) error {
	currentWait := baseWait
	currentMaxWait := maxWait

	for attempt := 0; attempt < maxRetries; attempt++ {
		token, err := c.getToken()
		if err != nil {
			return fmt.Errorf("error getting token: %w", err)
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("error marshaling request: %w", err)
		}

		reqURL := fmt.Sprintf("%s%s", c.vtexBaseURL, endpoint)
		req, err := http.NewRequest(method, reqURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Network error, retry with backoff
			time.Sleep(currentWait)
			currentWait = min(time.Duration(float64(currentWait)*adjustFactor), currentMaxWait)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		// Invalid or expired token - renew and retry
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			_, err := c.refreshToken()
			if err != nil {
				return fmt.Errorf("error refreshing token: %w", err)
			}
			continue
		}

		// Rate limit or temporary error (404, 504) - wait and retry
		if resp.StatusCode == 404 || resp.StatusCode == 504 || resp.StatusCode == 429 {
			time.Sleep(currentWait)
			currentWait = min(time.Duration(float64(currentWait)*adjustFactor), currentMaxWait)
			// Increase max wait slowly
			currentMaxWait = min(time.Duration(float64(currentMaxWait)*1.1), 15*time.Second)
			continue
		}

		// Server error (5xx) - retry
		if resp.StatusCode >= 500 {
			time.Sleep(currentWait)
			currentWait = min(time.Duration(float64(currentWait)*adjustFactor), currentMaxWait)
			continue
		}

		// Other error (4xx) - do not retry
		return fmt.Errorf("request failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	return fmt.Errorf("max retries (%d) exceeded", maxRetries)
}

// CreateUserRole creates a user with a role in VTEX
func (c *VtexClient) CreateUserRole(user UserRole) error {
	payload := UserRoleRequest{
		Users: []UserRole{user},
	}
	return c.doRequestWithRetry("POST", "/_v/create-user-role", payload)
}

// DeleteUserRole deletes a user with a role in VTEX
func (c *VtexClient) DeleteUserRole(user UserRole) error {
	payload := UserRoleRequest{
		Users: []UserRole{user},
	}
	return c.doRequestWithRetry("POST", "/_v/remove-user-role", payload)
}

// ReadUserRole checks if a user exists
// Note: VTEX does not have an endpoint to query users
// This makes the resource "write-only"
func (c *VtexClient) ReadUserRole(email, account, roleName string) (*UserRole, error) {
	// TODO: Implement if VTEX has an endpoint to query users
	return nil, nil
}
