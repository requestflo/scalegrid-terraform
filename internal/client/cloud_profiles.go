package client

import (
	"context"
	"net/http"
)

// ListCloudProfiles returns every cloud profile on the account.
func (c *Client) ListCloudProfiles(ctx context.Context) ([]CloudProfile, error) {
	var resp cloudProfileListResponse
	if err := c.do(ctx, http.MethodGet, "/cloud_profiles", nil, &resp); err != nil {
		return nil, err
	}
	return resp.CloudProfiles, nil
}

// GetCloudProfile fetches a single cloud profile by ID.
func (c *Client) GetCloudProfile(ctx context.Context, id string) (*CloudProfile, error) {
	var profile CloudProfile
	if err := c.do(ctx, http.MethodGet, "/cloud_profiles/"+id, nil, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// CreateCloudProfile registers cloud credentials for Bring Your Own Cloud
// deployments.
func (c *Client) CreateCloudProfile(ctx context.Context, profile CloudProfile) (*CloudProfile, error) {
	var created CloudProfile
	if err := c.do(ctx, http.MethodPost, "/cloud_profiles", profile, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateCloudProfile updates an existing cloud profile.
func (c *Client) UpdateCloudProfile(ctx context.Context, id string, profile CloudProfile) (*CloudProfile, error) {
	var updated CloudProfile
	if err := c.do(ctx, http.MethodPatch, "/cloud_profiles/"+id, profile, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

// DeleteCloudProfile removes a cloud profile.
func (c *Client) DeleteCloudProfile(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/cloud_profiles/"+id, nil, nil)
}
