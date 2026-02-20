package client

import (
	"bytes"
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
	err := c.makeRequest("login", params, &response)
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
	err := c.makeRequest("logout", params, &response)
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	c.sessionID = ""
	return nil
}

// makeRequest makes an HTTP request to the ISP Config API
func (c *Client) makeRequest(method string, params map[string]interface{}, result interface{}) error {
	// Build URL with method parameter
	apiURL := fmt.Sprintf("%s?%s", c.baseURL, method)

	// Convert params to JSON
	jsonData, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
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

// Web Domain methods

// AddWebDomain creates a new web domain
func (c *Client) AddWebDomain(domain *WebDomain, clientID int) (int, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"params":     domain,
	}

	var response APIResponse
	err := c.makeRequest("sites_web_domain_add", params, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to add web domain: %w", err)
	}

	if response.Code != "ok" {
		return 0, fmt.Errorf("failed to add web domain: %s", response.Message)
	}

	// Response should be the domain ID (can be float64 or string)
	if id, ok := response.Response.(float64); ok {
		return int(id), nil
	}
	if idStr, ok := response.Response.(string); ok {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return 0, fmt.Errorf("failed to parse domain ID string: %w", err)
		}
		return id, nil
	}

	return 0, fmt.Errorf("unexpected response type: %T", response.Response)
}

// GetWebDomain retrieves a web domain by ID
func (c *Client) GetWebDomain(domainID int) (*WebDomain, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": domainID,
	}

	var response APIResponse
	err := c.makeRequest("sites_web_domain_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get web domain: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get web domain: %s", response.Message)
	}

	// Parse the response into WebDomain
	jsonData, err := json.Marshal(response.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var domain WebDomain
	err = json.Unmarshal(jsonData, &domain)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal web domain: %w", err)
	}

	return &domain, nil
}

// UpdateWebDomain updates a web domain
func (c *Client) UpdateWebDomain(domainID int, clientID int, domain *WebDomain) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"primary_id": domainID,
		"params":     domain,
	}

	var response APIResponse
	err := c.makeRequest("sites_web_domain_update", params, &response)
	if err != nil {
		return fmt.Errorf("failed to update web domain: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to update web domain: %s", response.Message)
	}

	return nil
}

// DeleteWebDomain deletes a web domain
func (c *Client) DeleteWebDomain(domainID int) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": domainID,
	}

	var response APIResponse
	err := c.makeRequest("sites_web_domain_delete", params, &response)
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
func (c *Client) AddFTPUser(ftpUser *FTPUser, clientID int) (int, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"params":     ftpUser,
	}

	var response APIResponse
	err := c.makeRequest("sites_ftp_user_add", params, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to add FTP user: %w", err)
	}

	if response.Code != "ok" {
		return 0, fmt.Errorf("failed to add FTP user: %s", response.Message)
	}

	// Response should be the FTP user ID (can be float64 or string)
	if id, ok := response.Response.(float64); ok {
		return int(id), nil
	}
	if idStr, ok := response.Response.(string); ok {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return 0, fmt.Errorf("failed to parse FTP user ID string: %w", err)
		}
		return id, nil
	}

	return 0, fmt.Errorf("unexpected response type: %T", response.Response)
}

// GetFTPUser retrieves an FTP user by ID
func (c *Client) GetFTPUser(ftpUserID int) (*FTPUser, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": ftpUserID,
	}

	var response APIResponse
	err := c.makeRequest("sites_ftp_user_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get FTP user: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get FTP user: %s", response.Message)
	}

	jsonData, err := json.Marshal(response.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var ftpUser FTPUser
	err = json.Unmarshal(jsonData, &ftpUser)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal FTP user: %w", err)
	}

	return &ftpUser, nil
}

// UpdateFTPUser updates an FTP user
func (c *Client) UpdateFTPUser(ftpUserID int, clientID int, ftpUser *FTPUser) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"primary_id": ftpUserID,
		"params":     ftpUser,
	}

	var response APIResponse
	err := c.makeRequest("sites_ftp_user_update", params, &response)
	if err != nil {
		return fmt.Errorf("failed to update FTP user: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to update FTP user: %s", response.Message)
	}

	return nil
}

// DeleteFTPUser deletes an FTP user
func (c *Client) DeleteFTPUser(ftpUserID int) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": ftpUserID,
	}

	var response APIResponse
	err := c.makeRequest("sites_ftp_user_delete", params, &response)
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
func (c *Client) AddShellUser(shellUser *ShellUser, clientID int) (int, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"params":     shellUser,
	}

	var response APIResponse
	err := c.makeRequest("sites_shell_user_add", params, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to add shell user: %w", err)
	}

	if response.Code != "ok" {
		return 0, fmt.Errorf("failed to add shell user: %s", response.Message)
	}

	// Response should be the shell user ID (can be float64 or string)
	if id, ok := response.Response.(float64); ok {
		return int(id), nil
	}
	if idStr, ok := response.Response.(string); ok {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return 0, fmt.Errorf("failed to parse shell user ID string: %w", err)
		}
		return id, nil
	}

	return 0, fmt.Errorf("unexpected response type: %T", response.Response)
}

// GetShellUser retrieves a shell user by ID
func (c *Client) GetShellUser(shellUserID int) (*ShellUser, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": shellUserID,
	}

	var response APIResponse
	err := c.makeRequest("sites_shell_user_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get shell user: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get shell user: %s", response.Message)
	}

	// Parse the response into ShellUser
	jsonData, err := json.Marshal(response.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var shellUser ShellUser
	err = json.Unmarshal(jsonData, &shellUser)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal shell user: %w", err)
	}

	return &shellUser, nil
}

// UpdateShellUser updates a shell user
func (c *Client) UpdateShellUser(shellUserID int, clientID int, shellUser *ShellUser) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"primary_id": shellUserID,
		"params":     shellUser,
	}

	var response APIResponse
	err := c.makeRequest("sites_shell_user_update", params, &response)
	if err != nil {
		return fmt.Errorf("failed to update shell user: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to update shell user: %s", response.Message)
	}

	return nil
}

// DeleteShellUser deletes a shell user
func (c *Client) DeleteShellUser(shellUserID int) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": shellUserID,
	}

	var response APIResponse
	err := c.makeRequest("sites_shell_user_delete", params, &response)
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
func (c *Client) AddDatabase(database *Database, clientID int) (int, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"params":     database,
	}

	var response APIResponse
	err := c.makeRequest("sites_database_add", params, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to add database: %w", err)
	}

	if response.Code != "ok" {
		return 0, fmt.Errorf("failed to add database: %s", response.Message)
	}

	// Response should be the database ID (can be float64 or string)
	if id, ok := response.Response.(float64); ok {
		return int(id), nil
	}
	if idStr, ok := response.Response.(string); ok {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return 0, fmt.Errorf("failed to parse database ID string: %w", err)
		}
		return id, nil
	}

	return 0, fmt.Errorf("unexpected response type: %T", response.Response)
}

// GetDatabase retrieves a database by ID
func (c *Client) GetDatabase(databaseID int) (*Database, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": databaseID,
	}

	var response APIResponse
	err := c.makeRequest("sites_database_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get database: %s", response.Message)
	}

	jsonData, err := json.Marshal(response.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var database Database
	err = json.Unmarshal(jsonData, &database)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal database: %w", err)
	}

	return &database, nil
}

// UpdateDatabase updates a database
func (c *Client) UpdateDatabase(databaseID int, clientID int, database *Database) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"primary_id": databaseID,
		"params":     database,
	}

	var response APIResponse
	err := c.makeRequest("sites_database_update", params, &response)
	if err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to update database: %s", response.Message)
	}

	return nil
}

// DeleteDatabase deletes a database
func (c *Client) DeleteDatabase(databaseID int) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": databaseID,
	}

	var response APIResponse
	err := c.makeRequest("sites_database_delete", params, &response)
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
func (c *Client) AddDatabaseUser(dbUser *DatabaseUser, clientID int) (int, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"params":     dbUser,
	}

	var response APIResponse
	err := c.makeRequest("sites_database_user_add", params, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to add database user: %w", err)
	}

	if response.Code != "ok" {
		return 0, fmt.Errorf("failed to add database user: %s", response.Message)
	}

	// Response should be the database user ID (can be float64 or string)
	if id, ok := response.Response.(float64); ok {
		return int(id), nil
	}
	if idStr, ok := response.Response.(string); ok {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return 0, fmt.Errorf("failed to parse database user ID string: %w", err)
		}
		return id, nil
	}

	return 0, fmt.Errorf("unexpected response type: %T", response.Response)
}

// GetDatabaseUser retrieves a database user by ID
func (c *Client) GetDatabaseUser(dbUserID int) (*DatabaseUser, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": dbUserID,
	}

	var response APIResponse
	err := c.makeRequest("sites_database_user_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get database user: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get database user: %s", response.Message)
	}

	jsonData, err := json.Marshal(response.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var dbUser DatabaseUser
	err = json.Unmarshal(jsonData, &dbUser)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal database user: %w", err)
	}

	return &dbUser, nil
}

// UpdateDatabaseUser updates a database user
func (c *Client) UpdateDatabaseUser(dbUserID int, clientID int, dbUser *DatabaseUser) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
		"primary_id": dbUserID,
		"params":     dbUser,
	}

	var response APIResponse
	err := c.makeRequest("sites_database_user_update", params, &response)
	if err != nil {
		return fmt.Errorf("failed to update database user: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to update database user: %s", response.Message)
	}

	return nil
}

// DeleteDatabaseUser deletes a database user
func (c *Client) DeleteDatabaseUser(dbUserID int) error {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"primary_id": dbUserID,
	}

	var response APIResponse
	err := c.makeRequest("sites_database_user_delete", params, &response)
	if err != nil {
		return fmt.Errorf("failed to delete database user: %w", err)
	}

	if response.Code != "ok" {
		return fmt.Errorf("failed to delete database user: %s", response.Message)
	}

	return nil
}

// Server methods

// GetPHPVersions retrieves available PHP versions for a given server and PHP handler type.
// The phpType parameter should be "php-fpm", "fast-cgi", or "hhvm".
// Returns a map of short PHP version string -> full info string
// (e.g. "8.4" -> "PHP 8.4:/etc/init.d/php8.4-fpm:/etc/php/8.4/fpm:/etc/php/8.4/fpm/pool.d").
func (c *Client) GetPHPVersions(serverID int, phpType string) (map[string]string, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"server_id":  serverID,
		"php":        phpType,
	}

	var response APIResponse
	err := c.makeRequest("server_get_php_versions", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get PHP versions: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get PHP versions: %s", response.Message)
	}

	// Marshal response back to JSON for flexible parsing
	jsonData, err := json.Marshal(response.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PHP versions response: %w", err)
	}

	// The API returns a JSON array of PHP info strings, e.g.:
	//   ["PHP 7.0:/etc/init.d/php7.0-fpm:...", "PHP 8.4:..."]
	var phpVersionsList []string
	if err := json.Unmarshal(jsonData, &phpVersionsList); err != nil {
		return nil, fmt.Errorf("failed to parse PHP versions response: %w, body: %s", err, string(jsonData))
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
func (c *Client) GetClient(clientID int) (*ISPConfigClient, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
		"client_id":  clientID,
	}

	var response APIResponse
	err := c.makeRequest("client_get", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get client: %s", response.Message)
	}

	jsonData, err := json.Marshal(response.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var client ISPConfigClient
	err = json.Unmarshal(jsonData, &client)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal client: %w", err)
	}

	return &client, nil
}

// GetAllClients retrieves all clients
func (c *Client) GetAllClients() ([]ISPConfigClient, error) {
	params := map[string]interface{}{
		"session_id": c.getSessionID(),
	}

	var response APIResponse
	err := c.makeRequest("client_get_all", params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get all clients: %w", err)
	}

	if response.Code != "ok" {
		return nil, fmt.Errorf("failed to get all clients: %s", response.Message)
	}

	jsonData, err := json.Marshal(response.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var clients []ISPConfigClient
	err = json.Unmarshal(jsonData, &clients)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal clients: %w", err)
	}

	return clients, nil
}
