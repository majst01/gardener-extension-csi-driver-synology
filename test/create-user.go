package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// SynologyClient handles authentication and API calls to Synology DSM
type SynologyClient struct {
	BaseURL    string
	HTTPClient *http.Client
	SessionID  string
	SynoToken  string
	PublicKey  *rsa.PublicKey
}

// LoginResponse represents the login API response
type LoginResponse struct {
	Success bool `json:"success"`
	Data    struct {
		SID string `json:"sid"`
	} `json:"data"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code int `json:"code"`
	} `json:"error,omitempty"`
}

// EncryptionInfo represents the public key information
type EncryptionInfo struct {
	Success bool `json:"success"`
	Data    struct {
		PublicKey   string `json:"public_key"`
		CipherKey   string `json:"cipherkey"`
		CipherToken string `json:"ciphertoken"`
		ServerTime  int64  `json:"server_time"`
	} `json:"data"`
}

// PasswordCipher contains encrypted password data
type PasswordCipher struct {
	RSA string `json:"rsa"`
	AES string `json:"aes"`
}

// NewSynologyClient creates a new client instance
func NewSynologyClient(baseURL string) (*SynologyClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &SynologyClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Jar: jar,
		},
	}, nil
}

// GetEncryptionInfo retrieves the public key for password encryption
func (c *SynologyClient) GetEncryptionInfo() error {
	params := url.Values{}
	params.Set("api", "SYNO.API.Encryption")
	params.Set("method", "getinfo")
	params.Set("version", "1")
	params.Set("format", "module")

	resp, err := c.makeRequest("GET", "/webapi/encryption.cgi", params, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var encInfo EncryptionInfo
	if err := json.NewDecoder(resp.Body).Decode(&encInfo); err != nil {
		return err
	}

	if !encInfo.Success {
		return fmt.Errorf("failed to get encryption info")
	}

	// Parse the public key
	block, _ := pem.Decode([]byte(encInfo.Data.PublicKey))
	if block == nil {
		return fmt.Errorf("failed to parse PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	var ok bool
	c.PublicKey, ok = pub.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("not an RSA public key")
	}

	return nil
}

// encryptPassword encrypts the password using RSA+AES hybrid encryption
func (c *SynologyClient) encryptPassword(password string) (string, error) {
	// Generate a random AES key (32 bytes for AES-256)
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return "", err
	}

	// Encrypt the AES key with RSA
	rsaEncrypted, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		c.PublicKey,
		aesKey,
		nil,
	)
	if err != nil {
		return "", err
	}

	// Encrypt the password with AES-CBC
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	paddedPassword := pkcs7Pad([]byte(password), aes.BlockSize)

	// Generate IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := make([]byte, len(paddedPassword))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedPassword)

	// Combine IV and ciphertext (Salted__ prefix mimics OpenSSL format)
	saltedData := append([]byte("Salted__"), iv...)
	saltedData = append(saltedData, ciphertext...)

	// Create the cipher object
	cipher := PasswordCipher{
		RSA: base64.StdEncoding.EncodeToString(rsaEncrypted),
		AES: base64.StdEncoding.EncodeToString(saltedData),
	}

	cipherJSON, err := json.Marshal(cipher)
	if err != nil {
		return "", err
	}

	return string(cipherJSON), nil
}

// pkcs7Pad applies PKCS7 padding
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

// Login authenticates with the Synology DSM
func (c *SynologyClient) Login(username, password string) error {
	// Get encryption info first
	if err := c.GetEncryptionInfo(); err != nil {
		return fmt.Errorf("failed to get encryption info: %w", err)
	}

	// Encrypt password
	encryptedPwd, err := c.encryptPassword(password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	params := url.Values{}
	params.Set("api", "SYNO.API.Auth")
	params.Set("method", "login")
	params.Set("version", "7")
	params.Set("account", username)
	params.Set("passwd", encryptedPwd)
	params.Set("session", "FileStation")
	params.Set("format", "sid")

	resp, err := c.makeRequest("POST", "/webapi/auth.cgi", params, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return err
	}

	if !loginResp.Success {
		return fmt.Errorf("login failed")
	}

	c.SessionID = loginResp.Data.SID

	// Extract SYNO token from response headers or cookies
	// In practice, this might come from cookies or a separate API call
	c.SynoToken = c.extractSynoToken(resp)

	return nil
}

// extractSynoToken extracts the SYNO token from the response
func (c *SynologyClient) extractSynoToken(resp *http.Response) string {
	// The SYNO token is typically in the X-SYNO-TOKEN header or can be extracted
	// from the session. For this example, we'll show a placeholder.
	// In practice, you'd need to make an additional call or parse cookies.
	return "YOUR_SYNO_TOKEN" // This needs to be obtained from the actual session
}

// CreateUser creates a new user on Synology DSM
func (c *SynologyClient) CreateUser(username, password, description, email string) error {
	// First, encrypt the password
	encryptedPwd, err := c.encryptPassword(password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Build the compound request structure
	compound := []map[string]interface{}{
		{
			"api":                 "SYNO.Core.User",
			"method":              "create",
			"version":             1,
			"name":                username,
			"description":         description,
			"email":               email,
			"cannot_chg_passwd":   false,
			"expired":             "normal",
			"passwd_never_expire": true,
			"notify_by_email":     false,
			"__cIpHeRtExT":        encryptedPwd,
		},
	}

	compoundJSON, err := json.Marshal(compound)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("api", "SYNO.Entry.Request")
	params.Set("method", "request")
	params.Set("version", "1")
	params.Set("stop_when_error", "false")
	params.Set("mode", `"sequential"`)
	params.Set("compound", string(compoundJSON))
	params.Set("_sid", c.SessionID)

	resp, err := c.makeRequest("POST", "/webapi/entry.cgi", params, map[string]string{
		"X-SYNO-TOKEN": c.SynoToken,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}

	if !apiResp.Success {
		if apiResp.Error != nil {
			return fmt.Errorf("API error code: %d", apiResp.Error.Code)
		}
		return fmt.Errorf("user creation failed")
	}

	return nil
}

// ListUsers retrieves the list of users
func (c *SynologyClient) ListUsers() ([]byte, error) {
	params := url.Values{}
	params.Set("api", "SYNO.Core.User")
	params.Set("method", "list")
	params.Set("version", "1")
	params.Set("type", `"local"`)
	params.Set("offset", "0")
	params.Set("limit", "-1")
	params.Set("additional", `["email","description","expired","2fa_status"]`)
	params.Set("_sid", c.SessionID)

	resp, err := c.makeRequest("POST", "/webapi/entry.cgi", params, map[string]string{
		"X-SYNO-TOKEN": c.SynoToken,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// makeRequest is a helper to make HTTP requests
func (c *SynologyClient) makeRequest(method, endpoint string, params url.Values, headers map[string]string) (*http.Response, error) {
	var req *http.Request
	var err error

	fullURL := c.BaseURL + endpoint

	if method == "GET" {
		if params != nil {
			fullURL += "?" + params.Encode()
		}
		req, err = http.NewRequest(method, fullURL, nil)
	} else {
		req, err = http.NewRequest(method, fullURL, strings.NewReader(params.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	}

	if err != nil {
		return nil, err
	}

	// Add custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	req.Header.Set("User-Agent", "Go-Synology-Client/1.0")

	return c.HTTPClient.Do(req)
}

func main() {
	// Initialize the client
	client, err := NewSynologyClient("http://localhost:5000")
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	// Login
	fmt.Println("Logging in...")
	err = client.Login("admin", "your_admin_password")
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}
	fmt.Println("Login successful!")

	// Create a new user
	fmt.Println("Creating user...")
	err = client.CreateUser(
		"testuser",
		"SecurePassword123!",
		"Test User",
		"testuser@email.com",
	)
	if err != nil {
		fmt.Printf("User creation failed: %v\n", err)
		return
	}
	fmt.Println("User created successfully!")

	// List users
	fmt.Println("Listing users...")
	users, err := client.ListUsers()
	if err != nil {
		fmt.Printf("Failed to list users: %v\n", err)
		return
	}
	fmt.Printf("Users: %s\n", string(users))
}
