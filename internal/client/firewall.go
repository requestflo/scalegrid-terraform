package client

import (
	"context"
	"net/http"
)

// ListFirewallRules returns the firewall rules attached to a cluster.
func (c *Client) ListFirewallRules(ctx context.Context, clusterID string) ([]FirewallRule, error) {
	var resp firewallRuleListResponse
	if err := c.do(ctx, http.MethodGet, "/clusters/"+clusterID+"/firewall_rules", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Rules, nil
}

// GetFirewallRule fetches a single firewall rule by ID.
func (c *Client) GetFirewallRule(ctx context.Context, clusterID, id string) (*FirewallRule, error) {
	var rule FirewallRule
	if err := c.do(ctx, http.MethodGet, "/clusters/"+clusterID+"/firewall_rules/"+id, nil, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

// CreateFirewallRule adds a CIDR allow rule to a cluster.
func (c *Client) CreateFirewallRule(ctx context.Context, clusterID string, rule FirewallRule) (*FirewallRule, error) {
	var created FirewallRule
	if err := c.do(ctx, http.MethodPost, "/clusters/"+clusterID+"/firewall_rules", rule, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateFirewallRule modifies an existing firewall rule.
func (c *Client) UpdateFirewallRule(ctx context.Context, clusterID, id string, rule FirewallRule) (*FirewallRule, error) {
	var updated FirewallRule
	if err := c.do(ctx, http.MethodPatch, "/clusters/"+clusterID+"/firewall_rules/"+id, rule, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

// DeleteFirewallRule removes a firewall rule from a cluster.
func (c *Client) DeleteFirewallRule(ctx context.Context, clusterID, id string) error {
	return c.do(ctx, http.MethodDelete, "/clusters/"+clusterID+"/firewall_rules/"+id, nil, nil)
}
