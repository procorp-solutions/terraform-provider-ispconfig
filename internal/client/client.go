package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Client represents an ISP Config API client
type Client struct {
	baseURL    string
	username   string
	password   string
	sessionID  string
	httpClient *http.Client
	mu         sync.RWMutex
}

// NewClient creates a new ISP Config API client
func NewClient(host, username, password string, insecure bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	return &Client{
		baseURL:  fmt.Sprintf("https://%s/remote/json.php", host),
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// Login authenticates with the ISP Config API and stores the session ID
func (c *Client) Login() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	params := map[string]interface{}{
		"username": c.username,
		"password": c.password,
	}

	var response LoginResponse
	err := c.makeRequest(context.Background(), "login", params, &response)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("login failed: %s", response.Message)
	}

	// Extract session ID from response (should be a string on success)
	if sessionID, ok := response.Response.(string); ok {
		c.sessionID = sessionID
		return nil
	}

	return fmt.Errorf("login failed: unexpected response type: %T", response.Response)
}

// Logout closes the session with the ISP Config API
func (c *Client) Logout() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sessionID == "" {
		return nil
	}

	params := map[string]interface{}{
		"session_id": c.sessionID,
	}

	var response APIResponse
	err := c.makeRequest(context.Background(), "logout", params, &response)
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	c.sessionID = ""
	return nil
}

// makeRequest makes an HTTP request to the ISP Config API
func (c *Client) makeRequest(ctx context.Context, method string, params map[string]interface{}, result interface{}) error {
	// Build URL with method parameter
	apiURL := fmt.Sprintf("%s?%s", c.baseURL, method)

	// Convert params to JSON
	jsonData, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	err = json.Unmarshal(body, result)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	return nil
}

// getSessionID returns the current session ID
func (c *Client) getSessionID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionID
}

// parseResponseID extracts an integer ID from an API response that may be
// a float64 (JSON number) or a string.
func parseResponseID(response interface{}) (int, error) {
	if id, ok := response.(float64); ok {
		return int(id), nil
	}
	if idStr, ok := response.(string); ok {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return 0, fmt.Errorf("failed to parse ID string %q: %w", idStr, err)
		}
		return id, nil
	}
	return 0, fmt.Errorf("unexpected response type for ID: %T", response)
}

// unmarshalResponse re-marshals response.Response (an interface{}) into the
// concrete target struct via JSON round-trip.
func unmarshalResponse(response interface{}, target interface{}) error {
	jsonData, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}

// Web Domain methods

// AddWebDomain creates a new web domain
func (c *Client) AddWebDomain(ctx context.Context, domain *WebDomain, clientID int) (int, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"params":     domain,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_web_domain_add", params, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to add web domain: %w", err)
	}

	if response.Code != "ok" {
		return 0, fmt.Errorf("failed to add web domain: %s", response.Message)
	}

	return parseResponseID(response.Response)
}

// GetWebDomain retrieves a web domain by ID
func (c *Client) GetWebDomain(ctx context.Context, domainID int) (*WebDomain, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": domainID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_web_domain_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get web domain: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get web domain: %s", response.Message)
	}

	var domain WebDomain
	if err := unmarshalResponse(response.Response, &domain); err != nil {
		return nil, fmt.Errorf("failed to unmarshal web domain: %w", err)
	}

	return &domain, nil
}

// UpdateWebDomain updates a web domain
func (c *Client) UpdateWebDomain(ctx context.Context, domainID int, clientID int, domain *WebDomain) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"primary_id": domainID,
		"params":     domain,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_web_domain_update", params, &response)
	if err != nil {
		return fmt.Errorf("failed to update web domain: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to update web domain: %s", response.Message)
	}

	return nil
}

// DeleteWebDomain deletes a web domain
func (c *Client) DeleteWebDomain(ctx context.Context, domainID int) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": domainID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_web_domain_delete", params, &response)
	if err != nil {
		return fmt.Errorf("failed to delete web domain: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to delete web domain: %s", response.Message)
	}

	return nil
}

// FTP User methods

// AddFTPUser creates a new FTP user
func (c *Client) AddFTPUser(ctx context.Context, ftpUser *FTPUser, clientID int) (int, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"params":     ftpUser,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_ftp_user_add", params, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to add FTP user: %w", err)
	}

	if response.Code != "ok" {
		return 0, fmt.Errorf("failed to add FTP user: %s", response.Message)
	}

	return parseResponseID(response.Response)
}

// GetFTPUser retrieves an FTP user by ID
func (c *Client) GetFTPUser(ctx context.Context, ftpUserID int) (*FTPUser, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": ftpUserID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_ftp_user_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get FTP user: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get FTP user: %s", response.Message)
	}

	var ftpUser FTPUser
	if err := unmarshalResponse(response.Response, &ftpUser); err != nil {
		return nil, fmt.Errorf("failed to unmarshal FTP user: %w", err)
	}

	return &ftpUser, nil
}

// UpdateFTPUser updates an FTP user
func (c *Client) UpdateFTPUser(ctx context.Context, ftpUserID int, clientID int, ftpUser *FTPUser) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"primary_id": ftpUserID,
		"params":     ftpUser,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_ftp_user_update", params, &response)
	if err != nil {
		return fmt.Errorf("failed to update FTP user: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to update FTP user: %s", response.Message)
	}

	return nil
}

// DeleteFTPUser deletes an FTP user
func (c *Client) DeleteFTPUser(ctx context.Context, ftpUserID int) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": ftpUserID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_ftp_user_delete", params, &response)
	if err != nil {
		return fmt.Errorf("failed to delete FTP user: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to delete FTP user: %s", response.Message)
	}

	return nil
}

// Shell User methods

// AddShellUser creates a new shell user
func (c *Client) AddShellUser(ctx context.Context, shellUser *ShellUser, clientID int) (int, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"params":     shellUser,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_shell_user_add", params, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to add shell user: %w", err)
	}

	if response.Code != "ok" {
		return 0, fmt.Errorf("failed to add shell user: %s", response.Message)
	}

	return parseResponseID(response.Response)
}

// GetShellUser retrieves a shell user by ID
func (c *Client) GetShellUser(ctx context.Context, shellUserID int) (*ShellUser, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": shellUserID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_shell_user_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get shell user: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get shell user: %s", response.Message)
	}

	var shellUser ShellUser
	if err := unmarshalResponse(response.Response, &shellUser); err != nil {
		return nil, fmt.Errorf("failed to unmarshal shell user: %w", err)
	}

	return &shellUser, nil
}

// UpdateShellUser updates a shell user
func (c *Client) UpdateShellUser(ctx context.Context, shellUserID int, clientID int, shellUser *ShellUser) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"primary_id": shellUserID,
		"params":     shellUser,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_shell_user_update", params, &response)
	if err != nil {
		return fmt.Errorf("failed to update shell user: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to update shell user: %s", response.Message)
	}

	return nil
}

// DeleteShellUser deletes a shell user
func (c *Client) DeleteShellUser(ctx context.Context, shellUserID int) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": shellUserID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_shell_user_delete", params, &response)
	if err != nil {
		return fmt.Errorf("failed to delete shell user: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to delete shell user: %s", response.Message)
	}

	return nil
}

// Database methods

// AddDatabase creates a new database
func (c *Client) AddDatabase(ctx context.Context, database *Database, clientID int) (int, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"params":     database,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_database_add", params, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to add database: %w", err)
	}

	if response.Code != "ok" {
		return 0, fmt.Errorf("failed to add database: %s", response.Message)
	}

	return parseResponseID(response.Response)
}

// GetDatabase retrieves a database by ID
func (c *Client) GetDatabase(ctx context.Context, databaseID int) (*Database, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": databaseID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_database_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get database: %s", response.Message)
	}

	var database Database
	if err := unmarshalResponse(response.Response, &database); err != nil {
		return nil, fmt.Errorf("failed to unmarshal database: %w", err)
	}

	return &database, nil
}

// UpdateDatabase updates a database
func (c *Client) UpdateDatabase(ctx context.Context, databaseID int, clientID int, database *Database) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"primary_id": databaseID,
		"params":     database,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_database_update", params, &response)
	if err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to update database: %s", response.Message)
	}

	return nil
}

// DeleteDatabase deletes a database
func (c *Client) DeleteDatabase(ctx context.Context, databaseID int) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": databaseID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_database_delete", params, &response)
	if err != nil {
		return fmt.Errorf("failed to delete database: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to delete database: %s", response.Message)
	}

	return nil
}

// Database User methods

// AddDatabaseUser creates a new database user
func (c *Client) AddDatabaseUser(ctx context.Context, dbUser *DatabaseUser, clientID int) (int, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"params":     dbUser,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_database_user_add", params, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to add database user: %w", err)
	}

	if response.Code != "ok" {
		return 0, fmt.Errorf("failed to add database user: %s", response.Message)
	}

	return parseResponseID(response.Response)
}

// GetDatabaseUser retrieves a database user by ID
func (c *Client) GetDatabaseUser(ctx context.Context, dbUserID int) (*DatabaseUser, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": dbUserID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_database_user_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get database user: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get database user: %s", response.Message)
	}

	var dbUser DatabaseUser
	if err := unmarshalResponse(response.Response, &dbUser); err != nil {
		return nil, fmt.Errorf("failed to unmarshal database user: %w", err)
	}

	return &dbUser, nil
}

// UpdateDatabaseUser updates a database user
func (c *Client) UpdateDatabaseUser(ctx context.Context, dbUserID int, clientID int, dbUser *DatabaseUser) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"primary_id": dbUserID,
		"params":     dbUser,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_database_user_update", params, &response)
	if err != nil {
		return fmt.Errorf("failed to update database user: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to update database user: %s", response.Message)
	}

	return nil
}

// DeleteDatabaseUser deletes a database user
func (c *Client) DeleteDatabaseUser(ctx context.Context, dbUserID int) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": dbUserID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_database_user_delete", params, &response)
	if err != nil {
		return fmt.Errorf("failed to delete database user: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to delete database user: %s", response.Message)
	}

	return nil
}

// Cron Job methods

// AddCronJob creates a new cron task
func (c *Client) AddCronJob(ctx context.Context, cronJob *CronJob, clientID int) (int, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"params":     cronJob,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_cron_add", params, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to add cron job: %w", err)
	}

	if response.Code != "ok" {
		return 0, fmt.Errorf("failed to add cron job: %s", response.Message)
	}

	return parseResponseID(response.Response)
}

// GetCronJob retrieves a cron job by ID
func (c *Client) GetCronJob(ctx context.Context, cronJobID int) (*CronJob, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"cron_id":    cronJobID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_cron_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get cron job: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get cron job: %s", response.Message)
	}

	jsonData, err := json.Marshal(response.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	// ISPConfig may return an array ([{...}] or []) instead of a plain object
	if len(jsonData) > 0 && jsonData[0] == '[' {
		var arr []json.RawMessage
		if err := json.Unmarshal(jsonData, &arr); err != nil || len(arr) == 0 {
			return nil, fmt.Errorf("cron job not found (id: %d)", cronJobID)
		}
		jsonData = arr[0]
	}

	var cronJob CronJob
	err = json.Unmarshal(jsonData, &cronJob)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cron job: %w", err)
	}

	return &cronJob, nil
}

// UpdateCronJob updates a cron job
func (c *Client) UpdateCronJob(ctx context.Context, cronJobID int, clientID int, cronJob *CronJob) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"cron_id":    cronJobID,
		"params":     cronJob,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_cron_update", params, &response)
	if err != nil {
		return fmt.Errorf("failed to update cron job: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to update cron job: %s", response.Message)
	}

	return nil
}

// DeleteCronJob deletes a cron job
func (c *Client) DeleteCronJob(ctx context.Context, cronJobID int) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"cron_id":    cronJobID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "sites_cron_delete", params, &response)
	if err != nil {
		return fmt.Errorf("failed to delete cron job: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to delete cron job: %s", response.Message)
	}

	return nil
}

// Mail Domain methods

// AddMailDomain creates a new mail domain
func (c *Client) AddMailDomain(ctx context.Context, mailDomain *MailDomain, clientID int) (int, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"params":     mailDomain,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "mail_domain_add", params, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to add mail domain: %w", err)
	}

	if response.Code != "ok" {
		return 0, fmt.Errorf("failed to add mail domain: %s", response.Message)
	}

	return parseResponseID(response.Response)
}

// GetMailDomain retrieves a mail domain by ID
func (c *Client) GetMailDomain(ctx context.Context, mailDomainID int) (*MailDomain, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": mailDomainID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "mail_domain_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get mail domain: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get mail domain: %s", response.Message)
	}

	var mailDomain MailDomain
	if err := unmarshalResponse(response.Response, &mailDomain); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mail domain: %w", err)
	}

	return &mailDomain, nil
}

// UpdateMailDomain updates a mail domain
func (c *Client) UpdateMailDomain(ctx context.Context, mailDomainID int, clientID int, mailDomain *MailDomain) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"primary_id": mailDomainID,
		"params":     mailDomain,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "mail_domain_update", params, &response)
	if err != nil {
		return fmt.Errorf("failed to update mail domain: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to update mail domain: %s", response.Message)
	}

	return nil
}

// DeleteMailDomain deletes a mail domain
func (c *Client) DeleteMailDomain(ctx context.Context, mailDomainID int) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": mailDomainID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "mail_domain_delete", params, &response)
	if err != nil {
		return fmt.Errorf("failed to delete mail domain: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to delete mail domain: %s", response.Message)
	}

	return nil
}

// Mail User methods

// AddMailUser creates a new mail user (mailbox)
func (c *Client) AddMailUser(ctx context.Context, mailUser *MailUser, clientID int) (int, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"params":     mailUser,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "mail_user_add", params, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to add mail user: %w", err)
	}

	if response.Code != "ok" {
		return 0, fmt.Errorf("failed to add mail user: %s", response.Message)
	}

	return parseResponseID(response.Response)
}

// GetMailUser retrieves a mail user by ID
func (c *Client) GetMailUser(ctx context.Context, mailUserID int) (*MailUser, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": mailUserID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "mail_user_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get mail user: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get mail user: %s", response.Message)
	}

	var mailUser MailUser
	if err := unmarshalResponse(response.Response, &mailUser); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mail user: %w", err)
	}

	return &mailUser, nil
}

// UpdateMailUser updates a mail user (mailbox)
func (c *Client) UpdateMailUser(ctx context.Context, mailUserID int, clientID int, mailUser *MailUser) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"primary_id": mailUserID,
		"params":     mailUser,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "mail_user_update", params, &response)
	if err != nil {
		return fmt.Errorf("failed to update mail user: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to update mail user: %s", response.Message)
	}

	return nil
}

// DeleteMailUser deletes a mail user (mailbox)
func (c *Client) DeleteMailUser(ctx context.Context, mailUserID int) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": mailUserID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "mail_user_delete", params, &response)
	if err != nil {
		return fmt.Errorf("failed to delete mail user: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to delete mail user: %s", response.Message)
	}

	return nil
}

// Server methods

// GetPHPVersions retrieves available PHP versions for a given server and PHP handler type.
// The phpType parameter should be "php-fpm", "fast-cgi", or "hhvm".
// Returns a map of short PHP version string -> full info string
// (e.g. "8.4" -> "PHP 8.4:/etc/init.d/php8.4-fpm:/etc/php/8.4/fpm:/etc/php/8.4/fpm/pool.d").
func (c *Client) GetPHPVersions(ctx context.Context, serverID int, phpType string) (map[string]string, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"server_id":  serverID,
		"php":        phpType,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "server_get_php_versions", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get PHP versions: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get PHP versions: %s", response.Message)
	}

	// The API returns a JSON array of PHP info strings, e.g.:
	//   ["PHP 7.0:/etc/init.d/php7.0-fpm:...", "PHP 8.4:..."]
	var phpVersionsList []string
	if err := unmarshalResponse(response.Response, &phpVersionsList); err != nil {
		return nil, fmt.Errorf("failed to parse PHP versions response: %w", err)
	}

	result := make(map[string]string, len(phpVersionsList))
	for _, info := range phpVersionsList {
		version := ParsePHPVersion(info)
		if version == "" {
			return nil, fmt.Errorf("failed to extract PHP version from %q", info)
		}
		result[version] = info
	}

	return result, nil
}

// ParsePHPVersion extracts the version number from a PHP info string.
// Input format: "PHP 8.4:/etc/init.d/php8.4-fpm:/etc/php/8.4/fpm:/etc/php/8.4/fpm/pool.d"
// Returns: "8.4"
func ParsePHPVersion(info string) string {
	// Split by ":" to get the name part, e.g. "PHP 8.4"
	parts := strings.SplitN(info, ":", 2)
	if len(parts) == 0 {
		return ""
	}
	name := strings.TrimSpace(parts[0])

	// Remove "PHP " prefix (case-insensitive)
	if len(name) > 4 && strings.EqualFold(name[:4], "PHP ") {
		return strings.TrimSpace(name[4:])
	}

	return ""
}

// Client methods

// GetClient retrieves a client by ID
func (c *Client) GetClient(ctx context.Context, clientID int) (*ISPConfigClient, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
	}

	var response APIResponse
	err := c.makeRequest(ctx, "client_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get client: %s", response.Message)
	}

	var ispClient ISPConfigClient
	if err := unmarshalResponse(response.Response, &ispClient); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client: %w", err)
	}

	return &ispClient, nil
}

// GetAllClients retrieves all clients
func (c *Client) GetAllClients(ctx context.Context) ([]ISPConfigClient, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
	}

	var response APIResponse
	err := c.makeRequest(ctx, "client_get_all", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get all clients: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get all clients: %s", response.Message)
	}

	var clients []ISPConfigClient
	if err := unmarshalResponse(response.Response, &clients); err != nil {
		return nil, fmt.Errorf("failed to unmarshal clients: %w", err)
	}

	return clients, nil
}
