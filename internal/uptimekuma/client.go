package uptimekuma

import (
	"context"
	"fmt"

	"github.com/builtbyrobben/uptime-kuma-cli/internal/api"
)

type Client struct {
	api  *api.Client
	auth *AuthManager
}

func NewClient(baseURL, username, password string, opts ...api.ClientOption) *Client {
	auth := NewAuthManager(baseURL, username, password)

	allOpts := []api.ClientOption{
		api.WithUserAgent("uptime-kuma-cli/1.0"),
		api.WithAuthFn(auth.AuthFn()),
	}
	allOpts = append(allOpts, opts...)

	return &Client{
		api:  api.NewClient(baseURL, allOpts...),
		auth: auth,
	}
}

func (c *Client) ListMonitors(ctx context.Context) ([]Monitor, error) {
	var result []Monitor

	if err := c.api.Get(ctx, "/api/monitors", nil, &result); err != nil {
		return nil, fmt.Errorf("list monitors: %w", err)
	}

	return result, nil
}

func (c *Client) GetMonitor(ctx context.Context, id int) (*Monitor, error) {
	var result Monitor

	if err := c.api.Get(ctx, fmt.Sprintf("/api/monitors/%d", id), nil, &result); err != nil {
		return nil, fmt.Errorf("get monitor: %w", err)
	}

	return &result, nil
}

func (c *Client) CreateMonitor(ctx context.Context, input CreateMonitorInput) (*Monitor, error) {
	var result Monitor

	if err := c.api.Post(ctx, "/api/monitors", input, &result); err != nil {
		return nil, fmt.Errorf("create monitor: %w", err)
	}

	return &result, nil
}

func (c *Client) GetHeartbeats(ctx context.Context, id int) ([]Heartbeat, error) {
	var result []Heartbeat

	if err := c.api.Get(ctx, fmt.Sprintf("/api/monitors/%d/beats", id), nil, &result); err != nil {
		return nil, fmt.Errorf("get heartbeats: %w", err)
	}

	return result, nil
}

func (c *Client) PauseMonitor(ctx context.Context, id int) error {
	if err := c.api.Post(ctx, fmt.Sprintf("/api/monitors/%d/pause", id), nil, nil); err != nil {
		return fmt.Errorf("pause monitor: %w", err)
	}

	return nil
}

func (c *Client) ResumeMonitor(ctx context.Context, id int) error {
	if err := c.api.Post(ctx, fmt.Sprintf("/api/monitors/%d/resume", id), nil, nil); err != nil {
		return fmt.Errorf("resume monitor: %w", err)
	}

	return nil
}

func (c *Client) DeleteMonitor(ctx context.Context, id int) error {
	if err := c.api.Delete(ctx, fmt.Sprintf("/api/monitors/%d", id), nil); err != nil {
		return fmt.Errorf("delete monitor: %w", err)
	}

	return nil
}

func (c *Client) ListStatusPages(ctx context.Context) ([]StatusPage, error) {
	var result []StatusPage

	if err := c.api.Get(ctx, "/api/status-pages", nil, &result); err != nil {
		return nil, fmt.Errorf("list status pages: %w", err)
	}

	return result, nil
}

func (c *Client) GetStatusPage(ctx context.Context, slug string) (*StatusPage, error) {
	var result StatusPage

	if err := c.api.Get(ctx, fmt.Sprintf("/api/status-pages/%s", slug), nil, &result); err != nil {
		return nil, fmt.Errorf("get status page: %w", err)
	}

	return &result, nil
}

func (c *Client) Health(ctx context.Context) (*HealthStatus, error) {
	var result HealthStatus

	if err := c.api.Get(ctx, "/api/health", nil, &result); err != nil {
		return nil, fmt.Errorf("health check: %w", err)
	}

	return &result, nil
}
