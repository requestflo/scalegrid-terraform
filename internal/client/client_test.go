package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c, err := NewClient(Config{
		BaseURL:  srv.URL,
		Email:    "user@example.com",
		APIKey:   "secret",
		AuthMode: AuthBasic,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestNewClientValidation(t *testing.T) {
	if _, err := NewClient(Config{}); err == nil {
		t.Fatal("expected error when API key is missing")
	}
	if _, err := NewClient(Config{APIKey: "k", AuthMode: AuthBasic}); err == nil {
		t.Fatal("expected error when email is missing for basic auth")
	}
	if _, err := NewClient(Config{APIKey: "k", AuthMode: AuthBearer}); err != nil {
		t.Fatalf("bearer auth should not require email: %v", err)
	}
}

func TestBasicAuthHeader(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "user@example.com" || pass != "secret" {
			t.Errorf("unexpected basic auth: user=%q pass=%q ok=%v", user, pass, ok)
		}
		_ = json.NewEncoder(w).Encode(Cluster{ID: "c1", Name: "n"})
	})

	if _, err := c.GetCluster(context.Background(), "c1"); err != nil {
		t.Fatalf("GetCluster: %v", err)
	}
}

func TestBearerAuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token123" {
			t.Errorf("unexpected Authorization header: %q", got)
		}
		_ = json.NewEncoder(w).Encode(Cluster{ID: "c1"})
	}))
	defer srv.Close()

	c, err := NewClient(Config{BaseURL: srv.URL, APIKey: "token123", AuthMode: AuthBearer})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.GetCluster(context.Background(), "c1"); err != nil {
		t.Fatalf("GetCluster: %v", err)
	}
}

func TestCreateClusterRoundTrip(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/clusters" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var req ClusterCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Name != "prod" || req.DatabaseType != DatabaseMongoDB {
			t.Errorf("unexpected payload: %+v", req)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(asyncResponse{
			Cluster: Cluster{ID: "abc", Name: req.Name, DatabaseType: req.DatabaseType},
		})
	})

	cluster, err := c.CreateCluster(context.Background(), ClusterCreateRequest{
		Name:           "prod",
		DatabaseType:   DatabaseMongoDB,
		SizeID:         "small",
		CloudProfileID: "cp1",
	})
	if err != nil {
		t.Fatalf("CreateCluster: %v", err)
	}
	if cluster.ID != "abc" {
		t.Errorf("expected id abc, got %q", cluster.ID)
	}
}

func TestAPIErrorAndNotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "cluster not found"})
	})

	_, err := c.GetCluster(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true for: %v", err)
	}
}

func TestWaitForClusterReady(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		status := ClusterStatusProvisioning
		if calls >= 2 {
			status = ClusterStatusAvailable
		}
		_ = json.NewEncoder(w).Encode(Cluster{ID: "c1", Status: status})
	})

	cluster, err := c.WaitForClusterReady(context.Background(), "c1", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForClusterReady: %v", err)
	}
	if cluster.Status != ClusterStatusAvailable {
		t.Errorf("expected available, got %q", cluster.Status)
	}
	if calls < 2 {
		t.Errorf("expected at least 2 polls, got %d", calls)
	}
}

func TestListFirewallRules(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clusters/c1/firewall_rules" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(firewallRuleListResponse{
			Rules: []FirewallRule{{ID: "f1", CIDR: "10.0.0.0/8"}},
		})
	})

	rules, err := c.ListFirewallRules(context.Background(), "c1")
	if err != nil {
		t.Fatalf("ListFirewallRules: %v", err)
	}
	if len(rules) != 1 || rules[0].CIDR != "10.0.0.0/8" {
		t.Errorf("unexpected rules: %+v", rules)
	}
}
