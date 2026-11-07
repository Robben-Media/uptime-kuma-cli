package uptimekuma

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/builtbyrobben/uptime-kuma-cli/internal/api"
)

func TestClientCallerOptionsOverrideDefaults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/monitors" {
			t.Errorf("path = %s", r.URL.Path)
		}

		if r.Header.Get("User-Agent") != "selected-agent" || r.Header.Get("Authorization") != "Bearer selected" {
			t.Errorf("headers = %v", r.Header)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `[{"id":42,"name":"monitor"}]`)
	}))
	defer server.Close()
	client := NewClient(server.URL, "unused", "unused", api.WithUserAgent("selected-agent"), api.WithAuthFn(func(r *http.Request) { r.Header.Set("Authorization", "Bearer selected") }))

	monitors, err := client.ListMonitors(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(monitors) != 1 || monitors[0].ID != 42 {
		t.Fatalf("monitors = %#v", monitors)
	}
}
