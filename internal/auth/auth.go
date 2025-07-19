package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/yoshiken/vrc-print-upload/internal/config"
)

type Client struct {
	config     *config.Config
	httpClient *resty.Client
	cookies    map[string]*http.Cookie
}

type LoginOptions struct {
	Username     string
	Password     string
	RecoveryCode bool
}

type User struct {
	ID                   string `json:"id"`
	Username             string `json:"username"`
	DisplayName          string `json:"displayName"`
	TwoFactorAuthEnabled bool   `json:"twoFactorAuthEnabled"`
	RequiresTwoFactorAuth []string `json:"requiresTwoFactorAuth"`
}

type AuthResponse struct {
	User                 *User    `json:"user,omitempty"`
	RequiresTwoFactorAuth []string `json:"requiresTwoFactorAuth"`
	Error                string   `json:"error,omitempty"`
}

type TwoFactorAuthResponse struct {
	Verified bool   `json:"verified"`
	Error    string `json:"error,omitempty"`
}

func NewClient(cfg *config.Config) *Client {
	client := &Client{
		config:     cfg,
		httpClient: resty.New(),
		cookies:    make(map[string]*http.Cookie),
	}

	client.httpClient.SetBaseURL(cfg.APIBaseURL)
	client.httpClient.SetHeader("User-Agent", "vrc-print-upload/1.0")
	client.httpClient.OnAfterResponse(client.saveCookies)

	client.loadCookies()
	return client
}

func (c *Client) Login(opts LoginOptions) error {
	authHeader := c.createAuthHeader(opts.Username, opts.Password)
	
	resp, err := c.httpClient.R().
		SetHeader("Authorization", authHeader).
		SetResult(&AuthResponse{}).
		Get("/auth/user")

	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}

	authResp := resp.Result().(*AuthResponse)
	
	if authResp.Error != "" {
		return fmt.Errorf("authentication failed: %s", authResp.Error)
	}

	if len(authResp.RequiresTwoFactorAuth) > 0 || (authResp.User != nil && len(authResp.User.RequiresTwoFactorAuth) > 0) {
		return fmt.Errorf("2FA required - use VerifyTOTPCode or VerifyRecoveryCode methods")
	}

	if err := c.saveCookiesToFile(); err != nil {
		return fmt.Errorf("failed to save cookies: %w", err)
	}

	return nil
}


// VerifyTOTPCode verifies TOTP code programmatically (for GUI use)
func (c *Client) VerifyTOTPCode(code string) error {
	resp, err := c.httpClient.R().
		SetBody(map[string]string{"code": code}).
		SetResult(&TwoFactorAuthResponse{}).
		Post("/auth/twofactorauth/totp/verify")

	if err != nil {
		return fmt.Errorf("2FA verification failed: %w", err)
	}

	twoFAResp := resp.Result().(*TwoFactorAuthResponse)
	
	if !twoFAResp.Verified {
		return fmt.Errorf("2FA verification failed: invalid code")
	}

	if err := c.saveCookiesToFile(); err != nil {
		return fmt.Errorf("failed to save cookies: %w", err)
	}

	return nil
}


// VerifyRecoveryCode verifies recovery code programmatically (for GUI use)
func (c *Client) VerifyRecoveryCode(code string) error {
	resp, err := c.httpClient.R().
		SetBody(map[string]string{"code": code}).
		SetResult(&TwoFactorAuthResponse{}).
		Post("/auth/twofactorauth/recoverycode/verify")

	if err != nil {
		return fmt.Errorf("recovery code verification failed: %w", err)
	}

	twoFAResp := resp.Result().(*TwoFactorAuthResponse)
	
	if !twoFAResp.Verified {
		return fmt.Errorf("recovery code verification failed: invalid code")
	}

	if err := c.saveCookiesToFile(); err != nil {
		return fmt.Errorf("failed to save cookies: %w", err)
	}

	return nil
}

func (c *Client) IsAuthenticated() bool {
	authCookie, exists := c.cookies["auth"]
	if !exists || authCookie.Value == "" {
		return false
	}
	
	if !authCookie.Expires.IsZero() && authCookie.Expires.Before(time.Now()) {
		return false
	}
	
	return true
}

func (c *Client) GetCurrentUser() (*User, error) {
	resp, err := c.httpClient.R().
		SetResult(&User{}).
		Get("/auth/user")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", resp.Status())
	}

	return resp.Result().(*User), nil
}

func (c *Client) Logout() error {
	c.cookies = make(map[string]*http.Cookie)
	c.httpClient.SetCookies(nil)
	
	if err := os.Remove(c.config.CookieFile()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cookie file: %w", err)
	}
	
	return nil
}

func (c *Client) GetHTTPClient() *resty.Client {
	return c.httpClient
}

func (c *Client) createAuthHeader(username, password string) string {
	encodedUsername := url.QueryEscape(username)
	encodedPassword := url.QueryEscape(password)
	credentials := fmt.Sprintf("%s:%s", encodedUsername, encodedPassword)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encoded)
}

func (c *Client) saveCookies(client *resty.Client, resp *resty.Response) error {
	for _, cookie := range resp.Cookies() {
		c.cookies[cookie.Name] = cookie
	}
	
	var cookies []*http.Cookie
	for _, cookie := range c.cookies {
		cookies = append(cookies, cookie)
	}
	c.httpClient.SetCookies(cookies)
	
	return nil
}

func (c *Client) saveCookiesToFile() error {
	cookieFile := c.config.CookieFile()
	file, err := os.Create(cookieFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Set proper file permissions (owner read/write only)
	if err := os.Chmod(cookieFile, 0600); err != nil {
		return err
	}

	return json.NewEncoder(file).Encode(c.cookies)
}

func (c *Client) loadCookies() error {
	file, err := os.Open(c.config.CookieFile())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&c.cookies); err != nil {
		return err
	}

	var cookies []*http.Cookie
	for _, cookie := range c.cookies {
		cookies = append(cookies, cookie)
	}
	c.httpClient.SetCookies(cookies)

	return nil
}

