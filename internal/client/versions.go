package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
)

// GetDatabaseVersions returns the available database versions for an engine and
// cloud provider. cloudProvider should be one of AWS, AZURE, DO, or GCP.
func (c *Client) GetDatabaseVersions(ctx context.Context, db DBType, cloudProvider string) ([]string, error) {
	cloud, err := normalizeCloudProvider(cloudProvider)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("dbType", db.WireType())
	q.Set("cloudProvider", cloud)

	var resp databaseVersionsResponse
	path := "/Clusters/getDatabaseActiveVersions?" + q.Encode()
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Versions, nil
}

// versionList decodes the getDatabaseActiveVersions "versions" field. The API
// returns it either as a JSON array of version identifiers or as a JSON object
// mapping each version identifier to a display name; in both cases we expose the
// identifiers (the values accepted by a cluster's `version` attribute).
type versionList []string

func (v *versionList) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		*v = nil
		return nil
	}
	switch trimmed[0] {
	case '[':
		var arr []string
		if err := json.Unmarshal(trimmed, &arr); err != nil {
			return err
		}
		*v = arr
	case '{':
		var m map[string]json.RawMessage
		if err := json.Unmarshal(trimmed, &m); err != nil {
			return err
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		*v = keys
	default:
		return fmt.Errorf("scalegrid: unexpected versions payload %s", string(trimmed))
	}
	return nil
}

func normalizeCloudProvider(s string) (string, error) {
	switch s {
	case "AWS", "aws":
		return "EC2", nil
	case "AZURE", "azure":
		return "AZUREARM", nil
	case "DO", "do", "digitalocean", "DIGITALOCEAN":
		return "DIGITALOCEAN", nil
	case "GCP", "gcp", "google":
		return "GCP", nil
	default:
		return "", fmt.Errorf("scalegrid: unsupported cloud provider %q (use AWS, AZURE, DO, or GCP)", s)
	}
}
