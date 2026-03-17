package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient creates a Client pointing at the given test server.
func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	// Strip scheme for the client constructor, then override baseURL directly.
	c := &Client{
		baseURL:    server.URL + "/remote/json.php",
		username:   "admin",
		password:   "secret",
		sessionID:  "test-session",
		httpClient: server.Client(),
	}
	return c
}

// apiHandler returns an http.HandlerFunc that routes by the query-string method.
func apiHandler(routes map[string]func(params map[string]interface{}) interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method := r.URL.RawQuery // ISPConfig uses ?method_name as the query
		if method == "" {
			http.Error(w, "missing method", http.StatusBadRequest)
			return
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		handler, ok := routes[method]
		if !ok {
			resp := APIResponse{Code: "error", Message: "unknown method: " + method}
			json.NewEncoder(w).Encode(resp)
			return
		}

		response := handler(body)
		json.NewEncoder(w).Encode(APIResponse{Code: "ok", Message: "", Response: response})
	}
}

func TestLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["username"] != "admin" || body["password"] != "secret" {
			json.NewEncoder(w).Encode(LoginResponse{Code: "error", Message: "bad creds"})
			return
		}
		json.NewEncoder(w).Encode(LoginResponse{Code: "ok", Response: "session-abc"})
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL + "/remote/json.php",
		username:   "admin",
		password:   "secret",
		httpClient: server.Client(),
	}

	if err := c.Login(); err != nil {
		t.Fatalf("Login() error: %v", err)
	}
	if c.sessionID != "session-abc" {
		t.Errorf("sessionID = %q, want %q", c.sessionID, "session-abc")
	}
}

func TestLogin_BadCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(LoginResponse{Code: "error", Message: "invalid credentials"})
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL + "/remote/json.php",
		username:   "wrong",
		password:   "wrong",
		httpClient: server.Client(),
	}

	err := c.Login()
	if err == nil {
		t.Fatal("expected error for bad credentials, got nil")
	}
	if !strings.Contains(err.Error(), "invalid credentials") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "invalid credentials")
	}
}

func TestLogout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(APIResponse{Code: "ok"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	if err := c.Logout(); err != nil {
		t.Fatalf("Logout() error: %v", err)
	}
	if c.sessionID != "" {
		t.Errorf("sessionID = %q, want empty", c.sessionID)
	}
}

func TestAddWebDomain(t *testing.T) {
	server := httptest.NewServer(apiHandler(map[string]func(map[string]interface{}) interface{}{
		"sites_web_domain_add": func(params map[string]interface{}) interface{} {
			return float64(42) // API returns ID as number
		},
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx := context.Background()

	id, err := c.AddWebDomain(ctx, &WebDomain{Domain: "example.com"}, 1)
	if err != nil {
		t.Fatalf("AddWebDomain() error: %v", err)
	}
	if id != 42 {
		t.Errorf("got ID %d, want 42", id)
	}
}

func TestGetWebDomain(t *testing.T) {
	server := httptest.NewServer(apiHandler(map[string]func(map[string]interface{}) interface{}{
		"sites_web_domain_get": func(params map[string]interface{}) interface{} {
			return map[string]interface{}{
				"domain":    "example.com",
				"server_id": "1",
			}
		},
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx := context.Background()

	domain, err := c.GetWebDomain(ctx, 42)
	if err != nil {
		t.Fatalf("GetWebDomain() error: %v", err)
	}
	if domain.Domain != "example.com" {
		t.Errorf("Domain = %q, want %q", domain.Domain, "example.com")
	}
}

func TestDeleteWebDomain(t *testing.T) {
	server := httptest.NewServer(apiHandler(map[string]func(map[string]interface{}) interface{}{
		"sites_web_domain_delete": func(params map[string]interface{}) interface{} {
			return true
		},
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx := context.Background()

	if err := c.DeleteWebDomain(ctx, 42); err != nil {
		t.Fatalf("DeleteWebDomain() error: %v", err)
	}
}

func TestAddDatabase(t *testing.T) {
	server := httptest.NewServer(apiHandler(map[string]func(map[string]interface{}) interface{}{
		"sites_database_add": func(params map[string]interface{}) interface{} {
			return "99" // API sometimes returns ID as string
		},
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx := context.Background()

	id, err := c.AddDatabase(ctx, &Database{DatabaseName: "testdb", Type: "mysql"}, 1)
	if err != nil {
		t.Fatalf("AddDatabase() error: %v", err)
	}
	if id != 99 {
		t.Errorf("got ID %d, want 99", id)
	}
}

func TestMakeRequest_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	var resp APIResponse
	err := c.makeRequest(context.Background(), "test", map[string]interface{}{}, &resp)
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status code: 500") {
		t.Errorf("error = %q, want to contain status code info", err.Error())
	}
}

func TestMakeRequest_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(APIResponse{Code: "ok"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	var resp APIResponse
	err := c.makeRequest(ctx, "test", map[string]interface{}{}, &resp)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}
