package synology

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"time"
)

// Client represents a Synology DSM API client
type Client struct {
	url        string
	username   string
	password   string
	sessionID  string
	httpClient *http.Client
}

// NewClient creates a new Synology API client
func NewClient(url, username, password string) *Client {
	return &Client{
		url:      url,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// baseURL returns the base URL for API calls
func (c *Client) baseURL() string {
	return fmt.Sprintf("%s/webapi", c.url)
}

// Login authenticates with the Synology NAS
func (c *Client) Login() error {
	params := url.Values{}
	params.Add("api", "SYNO.API.Auth")
	params.Add("version", "3")
	params.Add("method", "login")
	params.Add("account", c.username)
	params.Add("passwd", c.password)
	params.Add("session", "FileStation")
	params.Add("format", "sid")

	resp, err := c.httpClient.Get(c.baseURL() + "/auth.cgi?" + params.Encode())
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			SID string `json:"sid"`
		} `json:"data"`
		Error struct {
			Code int `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("login failed with error code: %d", result.Error.Code)
	}

	c.sessionID = result.Data.SID
	return nil
}

// CreateUser creates a new user on the Synology NAS
func (c *Client) CreateUser(username, password string) error {
	if c.sessionID == "" {
		if err := c.Login(); err != nil {
			return err
		}
	}

	params := url.Values{}
	params.Add("api", "SYNO.Core.User")
	params.Add("method", "create")
	params.Add("version", "1")
	params.Add("name", username)
	params.Add("password", password)
	params.Add("_sid", c.sessionID)

	resp, err := c.httpClient.PostForm(c.baseURL()+"/entry.cgi", params)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Success bool `json:"success"`
		Error   struct {
			Code int `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		// User already exists (error code 407)
		if result.Error.Code == 407 {
			return nil
		}
		return fmt.Errorf("create user failed with error code: %d", result.Error.Code)
	}

	return nil
}

// DeleteUser deletes a user from the Synology NAS
func (c *Client) DeleteUser(username string) error {
	if c.sessionID == "" {
		if err := c.Login(); err != nil {
			return err
		}
	}

	params := url.Values{}
	params.Add("api", "SYNO.Core.User")
	params.Add("method", "delete")
	params.Add("version", "1")
	params.Add("name", username)
	params.Add("_sid", c.sessionID)

	resp, err := c.httpClient.PostForm(c.baseURL()+"/entry.cgi", params)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Success bool `json:"success"`
		Error   struct {
			Code int `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		// Ignore error if user doesn't exist
		if result.Error.Code == 407 {
			return nil
		}
		return fmt.Errorf("delete user failed with error code: %d", result.Error.Code)
	}

	return nil
}

// Logout ends the session
func (c *Client) Logout() error {
	if c.sessionID == "" {
		return nil
	}

	params := url.Values{}
	params.Add("api", "SYNO.API.Auth")
	params.Add("version", "1")
	params.Add("method", "logout")
	params.Add("session", "FileStation")
	params.Add("_sid", c.sessionID)

	_, err := c.httpClient.Get(c.baseURL() + "/auth.cgi?" + params.Encode())
	c.sessionID = ""
	return err
}

// GenerateRandomPassword generates a random password
func GenerateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)

	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[num.Int64()]
	}

	return string(password), nil
}

// GenerateShootUsername generates a username for a shoot cluster
func GenerateShootUsername(shootName, shootNamespace string) string {
	return fmt.Sprintf("gardener-%s-%s", shootNamespace, shootName)
}
