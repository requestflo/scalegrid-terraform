package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProviderSchemaValid(t *testing.T) {
	p := New("test")()
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("provider schema diagnostics: %v", resp.Diagnostics)
	}
	if _, ok := resp.Schema.Attributes["api_key"]; !ok {
		t.Error("expected api_key attribute on provider schema")
	}
	if diags := resp.Schema.ValidateImplementation(context.Background()); diags.HasError() {
		t.Fatalf("provider schema invalid: %v", diags)
	}
}

func TestResourceSchemasValid(t *testing.T) {
	resources := []func() resource.Resource{
		NewClusterResource,
		NewFirewallRuleResource,
		NewCloudProfileResource,
		NewBackupResource,
	}
	for _, ctor := range resources {
		r := ctor()
		resp := &resource.SchemaResponse{}
		r.Schema(context.Background(), resource.SchemaRequest{}, resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("resource schema diagnostics: %v", resp.Diagnostics)
			continue
		}
		if diags := resp.Schema.ValidateImplementation(context.Background()); diags.HasError() {
			t.Errorf("resource schema invalid: %v", diags)
		}
	}
}

func TestDataSourceSchemasValid(t *testing.T) {
	sources := []func() datasource.DataSource{
		NewClusterDataSource,
		NewClustersDataSource,
		NewCloudProfileDataSource,
	}
	for _, ctor := range sources {
		ds := ctor()
		resp := &datasource.SchemaResponse{}
		ds.Schema(context.Background(), datasource.SchemaRequest{}, resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("data source schema diagnostics: %v", resp.Diagnostics)
			continue
		}
		if diags := resp.Schema.ValidateImplementation(context.Background()); diags.HasError() {
			t.Errorf("data source schema invalid: %v", diags)
		}
	}
}

func TestProviderImplementsInterface(t *testing.T) {
	var _ provider.Provider = New("test")()
}
