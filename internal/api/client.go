package api

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/tcl-eaglelab/tcl-smart-home-cli/internal/config"
)

// Client is the TCL IoT API client.
// Not goroutine-safe: concurrent use requires external synchronization.
type Client struct {
	accessToken  string
	refreshToken string
	tokenExpires int64
	httpClient   *http.Client
}

// NewClient creates a new API client from stored config.
func NewClient() (*Client, error) {
	cfg, err := config.MustLoad()
	if err != nil {
		return nil, err
	}
	return &Client{
		accessToken:  cfg.AccessToken,
		refreshToken: cfg.RefreshToken,
		tokenExpires: cfg.TokenExpires,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// getToken returns a valid access token, refreshing if needed.
const tokenRefreshWindowMs = int64(300_000) // 5 minutes before expiry

func (c *Client) getToken() (string, error) {
	now := time.Now().UnixMilli()

	if c.accessToken != "" && c.tokenExpires > now+tokenRefreshWindowMs {
		return c.accessToken, nil
	}
	return c.refreshAccessToken()
}

// refreshAccessToken refreshes the token using the refresh token.
func (c *Client) refreshAccessToken() (string, error) {
	if c.refreshToken == "" {
		return "", fmt.Errorf("cannot refresh token: missing refreshToken.\nRun \"tcl init\" to re-authorize")
	}

	// "refershToken" is the actual API path on the TCL auth server (not a typo in our code).
	// Note: TCL auth API requires appId/appSecret/tenantId as URL query parameters.
	// This is the only authentication method the server supports for token refresh.
	u, err := url.Parse(config.AuthURL + "/auth/auth/refershToken")
	if err != nil {
		return "", fmt.Errorf("failed to parse auth URL: %w", err)
	}
	q := u.Query()
	q.Set("appId", config.AppID)
	q.Set("appSecret", config.AppSecret)
	q.Set("tenantId", config.TenantID)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("refreshToken", c.refreshToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token refresh request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	const maxAuthResponseSize = 1 * 1024 * 1024 // 1MB
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxAuthResponseSize))
	if err != nil {
		return "", fmt.Errorf("failed to read token refresh response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("token refresh returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result apiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("token refresh response parse failed: %w", err)
	}

	if !codeMatches(result.Code, 1) {
		msg := result.Msg
		if msg == "" {
			msg = "Unknown error"
		}
		return "", fmt.Errorf("token refresh failed: %s", msg)
	}

	c.accessToken = result.AccessToken
	c.refreshToken = result.RefreshToken
	expiry, err := config.TokenExpiryMs(result.AccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to parse token expiry: %w", err)
	}
	c.tokenExpires = expiry

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config for token persistence: %v\n", err)
	} else {
		cfg.AccessToken = c.accessToken
		cfg.RefreshToken = c.refreshToken
		cfg.TokenExpires = c.tokenExpires
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to persist refreshed token: %v\n", err)
		}
	}

	return c.accessToken, nil
}

// request makes an authenticated API request.
func (c *Client) request(method, reqURL string, body interface{}) (json.RawMessage, error) {
	return c.requestWithRetry(method, reqURL, body, 0)
}

func (c *Client) requestWithRetry(method, reqURL string, body interface{}, retryCount int) (json.RawMessage, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("accessToken", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("t-store-uuid", "OPENCLAW")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	const maxResponseSize = 10 * 1024 * 1024 // 10MB
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result apiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !isSuccess(result) {
		if isTokenError(result) && retryCount < 1 {
			if _, err := c.refreshAccessToken(); err == nil {
				return c.requestWithRetry(method, reqURL, body, retryCount+1)
			}
		}
		msg := result.Message
		if msg == "" {
			msg = result.Msg
		}
		if msg == "" {
			msg = "Unknown error"
		}
		return nil, fmt.Errorf("API Error [%v]: %s", result.Code, msg)
	}

	return result.Data, nil
}

// codeMatches handles comparison of API response codes.
// TCL API returns the "code" field as either a JSON number or a JSON string
// depending on the endpoint, so we must handle both types.
func codeMatches(code interface{}, targets ...float64) bool {
	switch v := code.(type) {
	case float64:
		for _, t := range targets {
			if v == t {
				return true
			}
		}
	case string:
		for _, t := range targets {
			if v == fmt.Sprintf("%.0f", t) {
				return true
			}
		}
	}
	return false
}

func isSuccess(r apiResponse) bool {
	if r.Success {
		return true
	}
	return codeMatches(r.Code, 200, 1)
}

func isTokenError(r apiResponse) bool {
	if codeMatches(r.Code, 401) {
		return true
	}
	return strings.Contains(strings.ToLower(r.Msg), "token") ||
		strings.Contains(strings.ToLower(r.Message), "token")
}

func randomMsgID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// --- Public API Methods ---

// GetAllDevices returns all devices for the current user.
func (c *Client) GetAllDevices() ([]Device, error) {
	data, err := c.request("GET", config.IoTURL+"/v1/tclplus/user/user_devices/openclaw", nil)
	if err != nil {
		return nil, err
	}
	var devices []Device
	if err := json.Unmarshal(data, &devices); err != nil {
		return nil, fmt.Errorf("failed to parse devices: %w", err)
	}
	return devices, nil
}

// GetDeviceStatus returns the current property values for a device.
func (c *Client) GetDeviceStatus(deviceID string) (map[string]interface{}, error) {
	body := map[string]string{"deviceId": deviceID}
	data, err := c.request("POST", config.IoTURL+"/v1/thing/status/openclaw", body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Status map[string]interface{} `json:"status"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}
	if result.Status == nil {
		return make(map[string]interface{}), nil
	}
	return result.Status, nil
}

// GetDeviceModel returns the Thing Model for a product.
func (c *Client) GetDeviceModel(productKey string) (*TSLModel, error) {
	data, err := c.request("GET", config.IoTURL+"/tsl/"+url.PathEscape(productKey)+"/tsl/openclaw", nil)
	if err != nil {
		return nil, err
	}
	var model TSLModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to parse model: %w", err)
	}
	return &model, nil
}

// ControlDevice sends property commands to a device.
func (c *Client) ControlDevice(deviceID string, commands map[string]interface{}) (json.RawMessage, error) {
	body := map[string]interface{}{
		"msgId":   randomMsgID(),
		"params":  []interface{}{commands},
		"version": "1.0",
	}
	return c.request("POST", config.IoTURL+"/v1/control/property/"+url.PathEscape(deviceID)+"/openclaw", body)
}
