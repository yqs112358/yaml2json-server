package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func setUp() {
	listenAddr = "0.0.0.0"
	listenPort = 8080
	authKey = ""
	urlSubPath = "/"
}

func TestYaml2JsonFromBody(t *testing.T) {
	setUp()
	yamlData := `
foo: bar
baz:
  - qux
  - quux
`
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(yamlData))
	req.Header.Set("Content-Type", "application/x-yaml")
	rr := httptest.NewRecorder()

	httpHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedJSON := `{"baz":["qux","quux"],"foo":"bar"}`
	if strings.TrimSpace(rr.Body.String()) != expectedJSON {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expectedJSON)
	}
}

func TestYaml2JsonFromURL(t *testing.T) {
	setUp()
	// Start a test server to serve YAML content
	yamlContent := `
foo: bar
numbers:
  - 1
  - 2
`
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		_, _ = w.Write([]byte(yamlContent))
	}))
	defer testServer.Close()

	req := httptest.NewRequest("GET", "/?url="+testServer.URL, nil)
	rr := httptest.NewRecorder()

	httpHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedJSON := `{"foo":"bar","numbers":[1,2]}`
	if strings.TrimSpace(rr.Body.String()) != expectedJSON {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expectedJSON)
	}
}

func TestAuthKeyInURL(t *testing.T) {
	setUp()
	authKey = "secret"

	yamlData := `foo: bar`
	req := httptest.NewRequest("POST", "/?key=secret", bytes.NewBufferString(yamlData))
	req.Header.Set("Content-Type", "application/x-yaml")
	rr := httptest.NewRecorder()

	httpHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestHttpBasicAuth(t *testing.T) {
	setUp()
	authKey = "secret"

	yamlData := `foo: bar`
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(yamlData))
	req.Header.Set("Content-Type", "application/x-yaml")

	// Set Basic Auth header
	auth := ":" + authKey
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encodedAuth)

	rr := httptest.NewRecorder()

	httpHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestUnauthorized(t *testing.T) {
	setUp()
	authKey = "secret"

	yamlData := `foo: bar`
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(yamlData))
	req.Header.Set("Content-Type", "application/x-yaml")
	rr := httptest.NewRecorder()

	httpHandler(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	var resp map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["error"] != "Unauthorized" {
		t.Errorf("Expected Unauthorized error, got %v", resp["error"])
	}
}

func TestNoYAMLProvided(t *testing.T) {
	setUp()
	req := httptest.NewRequest("POST", "/", nil)
	rr := httptest.NewRecorder()

	httpHandler(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var resp map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["error"] != "No YAML to convert" {
		t.Errorf("Expected 'No YAML to convert' error, got %v", resp["error"])
	}
}

func TestInvalidYAML(t *testing.T) {
	setUp()
	invalidYAML := `foo: bar: baz`
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(invalidYAML))
	req.Header.Set("Content-Type", "application/x-yaml")
	rr := httptest.NewRecorder()

	httpHandler(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	var resp map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["error"] != "Failed to convert YAML to JSON" {
		t.Errorf("Expected 'Failed to convert YAML to JSON' error, got %v", resp["error"])
	}
}

func TestInvalidURLParameter(t *testing.T) {
	setUp()
	req := httptest.NewRequest("GET", "/?url=invalid-url", nil)
	rr := httptest.NewRecorder()

	httpHandler(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var resp map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["error"] != "Failed to fetch YAML from given URL" {
		t.Errorf("Expected 'Failed to fetch YAML from given URL' error, got %v", resp["error"])
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("Y2JS_LISTEN_ADDR", "127.0.0.1")
	_ = os.Setenv("Y2JS_LISTEN_PORT", "9090")
	_ = os.Setenv("Y2JS_AUTH_KEY", "envsecret")
	_ = os.Setenv("Y2JS_URL_SUB_PATH", "/convert")

	parseConfigs()

	if listenAddr != "127.0.0.1" {
		t.Errorf("Expected listenAddr to be '127.0.0.1', got %v", listenAddr)
	}
	if listenPort != 9090 {
		t.Errorf("Expected listenPort to be 9090, got %v", listenPort)
	}
	if authKey != "envsecret" {
		t.Errorf("Expected authKey to be 'envsecret', got %v", authKey)
	}
	if urlSubPath != "/convert" {
		t.Errorf("Expected urlSubPath to be '/convert', got %v", urlSubPath)
	}

	// Clean up environment variables
	_ = os.Unsetenv("Y2JS_LISTEN_ADDR")
	_ = os.Unsetenv("Y2JS_LISTEN_PORT")
	_ = os.Unsetenv("Y2JS_AUTH_KEY")
	_ = os.Unsetenv("Y2JS_URL_SUB_PATH")
}
