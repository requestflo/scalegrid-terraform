package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// GetJob fetches the status of an asynchronous job.
func (c *Client) GetJob(ctx context.Context, id string) (*Job, error) {
	var job Job
	if err := c.do(ctx, http.MethodGet, "/jobs/"+id, nil, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// Cluster lifecycle status values returned by the API.
const (
	ClusterStatusProvisioning = "provisioning"
	ClusterStatusAvailable    = "available"
	ClusterStatusRunning      = "running"
	ClusterStatusError        = "error"
	ClusterStatusDeleting     = "deleting"
)

// WaitForClusterReady polls a cluster until it reaches a terminal "ready"
// status (available/running) or the context is cancelled. The pollInterval
// controls how often the cluster is re-read.
func (c *Client) WaitForClusterReady(ctx context.Context, id string, pollInterval time.Duration) (*Cluster, error) {
	if pollInterval <= 0 {
		pollInterval = 15 * time.Second
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		cluster, err := c.GetCluster(ctx, id)
		if err != nil {
			return nil, err
		}

		switch cluster.Status {
		case ClusterStatusAvailable, ClusterStatusRunning, "":
			// An empty status is treated as ready: some endpoints do not
			// report a transitional status once provisioning completes.
			return cluster, nil
		case ClusterStatusError:
			return cluster, fmt.Errorf("scalegrid: cluster %s entered error state", id)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("scalegrid: timed out waiting for cluster %s to become ready: %w", id, ctx.Err())
		case <-ticker.C:
		}
	}
}

// WaitForClusterDeleted polls a cluster until the API reports it gone (404).
func (c *Client) WaitForClusterDeleted(ctx context.Context, id string, pollInterval time.Duration) error {
	if pollInterval <= 0 {
		pollInterval = 15 * time.Second
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		_, err := c.GetCluster(ctx, id)
		if IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("scalegrid: timed out waiting for cluster %s to delete: %w", id, ctx.Err())
		case <-ticker.C:
		}
	}
}
