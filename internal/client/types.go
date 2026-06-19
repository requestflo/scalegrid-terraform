package client

// DatabaseType enumerates the database engines ScaleGrid supports.
const (
	DatabaseMongoDB    = "mongodb"
	DatabaseRedis      = "redis"
	DatabaseMySQL      = "mysql"
	DatabasePostgreSQL = "postgresql"
)

// DeploymentType enumerates the cluster topologies ScaleGrid supports. Not
// every topology is valid for every database engine.
const (
	DeploymentStandalone = "standalone"
	DeploymentReplicaSet = "replicaset"
	DeploymentSharded    = "sharded"
	DeploymentCluster    = "cluster"
)

// Cluster represents a ScaleGrid database deployment.
type Cluster struct {
	ID               string `json:"id,omitempty"`
	Name             string `json:"name"`
	DatabaseType     string `json:"database_type"`
	Version          string `json:"version,omitempty"`
	DeploymentType   string `json:"deployment_type,omitempty"`
	CloudProfileID   string `json:"cloud_profile_id,omitempty"`
	Region           string `json:"region,omitempty"`
	SizeID           string `json:"size_id,omitempty"`
	DiskSizeGB       int64  `json:"disk_size_gb,omitempty"`
	ShardCount       int64  `json:"shard_count,omitempty"`
	SSLEnabled       bool   `json:"ssl_enabled"`
	EncryptionAtRest bool   `json:"encryption_at_rest"`

	// Computed / read-only fields.
	Status           string   `json:"status,omitempty"`
	Host             string   `json:"host,omitempty"`
	Port             int64    `json:"port,omitempty"`
	ConnectionString string   `json:"connection_string,omitempty"`
	Nodes            []Node   `json:"nodes,omitempty"`
	CreatedAt        string   `json:"created_at,omitempty"`
	UpdatedAt        string   `json:"updated_at,omitempty"`
	Tags             []string `json:"tags,omitempty"`
}

// Node describes a single machine that backs a cluster.
type Node struct {
	ID        string `json:"id,omitempty"`
	Hostname  string `json:"hostname,omitempty"`
	PrivateIP string `json:"private_ip,omitempty"`
	PublicIP  string `json:"public_ip,omitempty"`
	Role      string `json:"role,omitempty"`
	Region    string `json:"region,omitempty"`
	Status    string `json:"status,omitempty"`
}

// ClusterCreateRequest is the payload sent to create a cluster.
type ClusterCreateRequest struct {
	Name             string   `json:"name"`
	DatabaseType     string   `json:"database_type"`
	Version          string   `json:"version,omitempty"`
	DeploymentType   string   `json:"deployment_type,omitempty"`
	CloudProfileID   string   `json:"cloud_profile_id"`
	Region           string   `json:"region,omitempty"`
	SizeID           string   `json:"size_id"`
	DiskSizeGB       int64    `json:"disk_size_gb,omitempty"`
	ShardCount       int64    `json:"shard_count,omitempty"`
	SSLEnabled       bool     `json:"ssl_enabled"`
	EncryptionAtRest bool     `json:"encryption_at_rest"`
	Tags             []string `json:"tags,omitempty"`
}

// ClusterUpdateRequest carries the mutable fields of a cluster. Only a subset
// of cluster attributes can be changed in place; the rest force replacement.
type ClusterUpdateRequest struct {
	Name       string   `json:"name,omitempty"`
	SizeID     string   `json:"size_id,omitempty"`
	DiskSizeGB int64    `json:"disk_size_gb,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

// clusterListResponse wraps the list endpoint's response envelope.
type clusterListResponse struct {
	Clusters []Cluster `json:"clusters"`
}

// asyncResponse is returned by endpoints that kick off a background job.
type asyncResponse struct {
	JobID   string  `json:"job_id,omitempty"`
	Cluster Cluster `json:"cluster,omitempty"`
}

// FirewallRule allows traffic from a CIDR to a cluster.
type FirewallRule struct {
	ID          string `json:"id,omitempty"`
	ClusterID   string `json:"cluster_id,omitempty"`
	CIDR        string `json:"cidr"`
	Description string `json:"description,omitempty"`
}

type firewallRuleListResponse struct {
	Rules []FirewallRule `json:"firewall_rules"`
}

// CloudProfile stores the cloud credentials/configuration used to provision
// clusters in a customer's own cloud account (Bring Your Own Cloud).
type CloudProfile struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name"`
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region,omitempty"`

	// Provider-specific credential fields. Sensitive values are write-only and
	// are not returned by the API on read.
	AccessKey      string `json:"access_key,omitempty"`
	SecretKey      string `json:"secret_key,omitempty"`
	SubscriptionID string `json:"subscription_id,omitempty"`
	TenantID       string `json:"tenant_id,omitempty"`
	ClientID       string `json:"client_id,omitempty"`
	ClientSecret   string `json:"client_secret,omitempty"`

	CreatedAt string `json:"created_at,omitempty"`
}

type cloudProfileListResponse struct {
	CloudProfiles []CloudProfile `json:"cloud_profiles"`
}

// Backup represents a single backup of a cluster.
type Backup struct {
	ID        string `json:"id,omitempty"`
	ClusterID string `json:"cluster_id,omitempty"`
	Status    string `json:"status,omitempty"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
	Type      string `json:"type,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type backupListResponse struct {
	Backups []Backup `json:"backups"`
}

// Credentials holds connection credentials for a cluster.
type Credentials struct {
	Username         string `json:"username,omitempty"`
	Password         string `json:"password,omitempty"`
	Database         string `json:"database,omitempty"`
	ConnectionString string `json:"connection_string,omitempty"`
}

// Job represents an asynchronous operation in ScaleGrid.
type Job struct {
	ID       string `json:"id,omitempty"`
	Status   string `json:"status,omitempty"`
	Type     string `json:"type,omitempty"`
	Progress int64  `json:"progress,omitempty"`
	Message  string `json:"message,omitempty"`
}

// Job status values.
const (
	JobStatusQueued    = "queued"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
)
