package client

import (
	"context"
	"net/http"
)

// ListBackups returns the backups for a cluster.
func (c *Client) ListBackups(ctx context.Context, clusterID string) ([]Backup, error) {
	var resp backupListResponse
	if err := c.do(ctx, http.MethodGet, "/clusters/"+clusterID+"/backups", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Backups, nil
}

// GetBackup fetches a single backup by ID.
func (c *Client) GetBackup(ctx context.Context, clusterID, id string) (*Backup, error) {
	var backup Backup
	if err := c.do(ctx, http.MethodGet, "/clusters/"+clusterID+"/backups/"+id, nil, &backup); err != nil {
		return nil, err
	}
	return &backup, nil
}

// CreateBackup triggers an on-demand backup of a cluster.
func (c *Client) CreateBackup(ctx context.Context, clusterID string) (*Backup, error) {
	var backup Backup
	if err := c.do(ctx, http.MethodPost, "/clusters/"+clusterID+"/backups", struct{}{}, &backup); err != nil {
		return nil, err
	}
	return &backup, nil
}

// DeleteBackup removes a backup.
func (c *Client) DeleteBackup(ctx context.Context, clusterID, id string) error {
	return c.do(ctx, http.MethodDelete, "/clusters/"+clusterID+"/backups/"+id, nil, nil)
}
