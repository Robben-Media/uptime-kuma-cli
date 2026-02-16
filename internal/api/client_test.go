package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Authorization header 'Bearer test-token', got %s", r.Header.Get("Authorization"))
		}

		resp := map[string]any{
			"id":   1,
			"name": "Test Monitor",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, WithAuthFn(func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer test-token")
	}))

	var result struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	err := client.Get(context.Background(), "/api/monitors/1", nil, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}

	if result.Name != "Test Monitor" {
		t.Errorf("expected Name 'Test Monitor', got %s", result.Name)
	}
}

func TestPost_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		resp := map[string]any{
			"token": "jwt-test-token",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	var result struct {
		Token string `json:"token"`
	}

	err := client.Post(context.Background(), "/api/login", map[string]string{
		"username": "admin",
		"password": "secret",
	}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Token != "jwt-test-token" {
		t.Errorf("expected token 'jwt-test-token', got %s", result.Token)
	}
}

func TestGet_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
	}))
	defer server.Close()

	client := NewClient(server.URL)

	var result struct{}

	err := client.Get(context.Background(), "/api/monitors", nil, &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		resp := map[string]any{
			"msg": "Deleted Successfully.",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, WithAuthFn(func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer test-token")
	}))

	var result struct {
		Msg string `json:"msg"`
	}

	err := client.Delete(context.Background(), "/api/monitors/1", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Msg != "Deleted Successfully." {
		t.Errorf("expected msg 'Deleted Successfully.', got %s", result.Msg)
	}
}
