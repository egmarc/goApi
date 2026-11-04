package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/egmarc/goApi/internal/api"
)

func TestAPI(t *testing.T) {
	const token = "integration-test-token"
	handler, err := api.NewHandler(token)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	client := &http.Client{Timeout: 5 * time.Second}

	tests := []struct {
		name    string
		method  string
		path    string
		headers []string
		status  int
		body    map[string]string
	}{
		{name: "public health", method: "GET", path: "/healthz", status: http.StatusOK, body: map[string]string{"status": "ok"}},
		{name: "missing token", method: "GET", path: "/api/v1/protected", status: http.StatusUnauthorized},
		{name: "incorrect token", method: "GET", path: "/api/v1/protected", headers: []string{"Bearer" + " incorrect"}, status: http.StatusUnauthorized},
		{name: "wrong scheme", method: "GET", path: "/api/v1/protected", headers: []string{"Basic " + token}, status: http.StatusUnauthorized},
		{name: "empty credential", method: "GET", path: "/api/v1/protected", headers: []string{"Bearer "}, status: http.StatusUnauthorized},
		{name: "extra credentials", method: "GET", path: "/api/v1/protected", headers: []string{"Bearer " + token + " extra"}, status: http.StatusUnauthorized},
		{name: "multiple headers", method: "GET", path: "/api/v1/protected", headers: []string{"Bearer " + token, "Bearer" + " incorrect"}, status: http.StatusUnauthorized},
		{name: "authorized", method: "GET", path: "/api/v1/protected", headers: []string{"Bearer " + token}, status: http.StatusOK, body: map[string]string{"message": "authorized"}},
		{name: "case insensitive scheme", method: "GET", path: "/api/v1/protected", headers: []string{"bearer " + token}, status: http.StatusOK, body: map[string]string{"message": "authorized"}},
		{name: "head requires authorization", method: "HEAD", path: "/api/v1/protected", status: http.StatusUnauthorized},
		{name: "unsupported method", method: "POST", path: "/api/v1/protected", headers: []string{"Bearer " + token}, status: http.StatusMethodNotAllowed},
		{name: "unknown route", method: "GET", path: "/missing", status: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, server.URL+tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, header := range tt.headers {
				req.Header.Add("Authorization", header)
			}
			res, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != tt.status {
				t.Fatalf("status = %d, want %d", res.StatusCode, tt.status)
			}
			if tt.status == http.StatusUnauthorized && res.Header.Get("WWW-Authenticate") != "Bearer"+` realm="api"` {
				t.Fatal("missing bearer authentication challenge")
			}
			if tt.method == "HEAD" || (tt.body == nil && tt.status != http.StatusUnauthorized) {
				return
			}
			if res.Header.Get("Content-Type") != "application/json" {
				t.Fatal("expected a JSON response")
			}
			var body map[string]string
			if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			want := tt.body
			if tt.status == http.StatusUnauthorized {
				want = map[string]string{"error": "unauthorized"}
			}
			if len(body) != len(want) {
				t.Fatalf("body = %v, want %v", body, want)
			}
			for key, value := range want {
				if body[key] != value {
					t.Fatalf("body = %v, want %v", body, want)
				}
			}
		})
	}
}
