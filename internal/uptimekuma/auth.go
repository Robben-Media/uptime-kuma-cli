package uptimekuma

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/builtbyrobben/uptime-kuma-cli/internal/api"
)

type AuthManager struct {
	baseURL  string
	username string
	password string
	token    string
	mu       sync.Mutex
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func NewAuthManager(baseURL, username, password string) *AuthManager {
	return &AuthManager{
		baseURL:  baseURL,
		username: username,
		password: password,
	}
}

func (a *AuthManager) GetToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.token != "" {
		return a.token, nil
	}

	loginClient := api.NewClient(a.baseURL, api.WithUserAgent("uptime-kuma-cli/1.0"))

	var resp loginResponse

	err := loginClient.Post(ctx, "/api/login", loginRequest{
		Username: a.username,
		Password: a.password,
	}, &resp)
	if err != nil {
		return "", fmt.Errorf("login failed: %w", err)
	}

	a.token = resp.Token

	return a.token, nil
}

func (a *AuthManager) AuthFn() func(*http.Request) {
	return func(r *http.Request) {
		token, err := a.GetToken(r.Context())
		if err == nil && token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
	}
}
