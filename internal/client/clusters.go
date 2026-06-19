package client

import (
	"context"
	"fmt"
	"net/http"
)

// CreateCluster provisions a new database deployment. ScaleGrid provisioning is
// asynchronous: the returned Cluster carries an ID immediately, but callers
// should poll GetCluster (or WaitForClusterReady) until it becomes available.
func (c *Client) CreateCluster(ctx context.Context, req ClusterCreateRequest) (*Cluster, error) {
	var resp asyncResponse
	if err := c.do(ctx, http.MethodPost, "/clusters", req, &resp); err != nil {
		return nil, err
	}
	cluster := resp.Cluster
	if cluster.ID == "" {
		return nil, fmt.Errorf("scalegrid: create cluster response did not include an id")
	}
	return &cluster, nil
}

// GetCluster fetches a single cluster by ID.
func (c *Client) GetCluster(ctx context.Context, id string) (*Cluster, error) {
	var cluster Cluster
	if err := c.do(ctx, http.MethodGet, "/clusters/"+id, nil, &cluster); err != nil {
		return nil, err
	}
	return &cluster, nil
}

// ListClusters returns every cluster on the account.
func (c *Client) ListClusters(ctx context.Context) ([]Cluster, error) {
	var resp clusterListResponse
	if err := c.do(ctx, http.MethodGet, "/clusters", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Clusters, nil
}

// UpdateCluster applies in-place changes (rename, scale up/down, resize disk).
func (c *Client) UpdateCluster(ctx context.Context, id string, req ClusterUpdateRequest) (*Cluster, error) {
	var cluster Cluster
	if err := c.do(ctx, http.MethodPatch, "/clusters/"+id, req, &cluster); err != nil {
		return nil, err
	}
	return &cluster, nil
}

// DeleteCluster tears down a cluster.
func (c *Client) DeleteCluster(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/clusters/"+id, nil, nil)
}

// GetClusterCredentials returns the connection credentials for a cluster.
func (c *Client) GetClusterCredentials(ctx context.Context, id string) (*Credentials, error) {
	var creds Credentials
	if err := c.do(ctx, http.MethodGet, "/clusters/"+id+"/credentials", nil, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}
