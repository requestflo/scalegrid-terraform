package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/requestflo/scalegrid-terraform/internal/client"
)

// diagnosticsList is a short alias for the framework diagnostics type, used by
// helper functions that accumulate diagnostics.
type diagnosticsList = diag.Diagnostics

// tagsFromList converts a Terraform list of strings into a Go slice.
func tagsFromList(ctx context.Context, list types.List) ([]string, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var tags []string
	diags := list.ElementsAs(ctx, &tags, false)
	return tags, diags
}

// tagsToList converts a Go slice of strings into a Terraform list.
func tagsToList(ctx context.Context, tags []string) (types.List, diag.Diagnostics) {
	return types.ListValueFrom(ctx, types.StringType, tags)
}

// firstNonEmpty returns the first non-empty string from the supplied values.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// stringValue safely extracts a Go string from a types.String, returning "" for
// null or unknown values.
func stringValue(s types.String) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	return s.ValueString()
}

// clientFromProviderData performs the type assertion shared by every resource
// and data source Configure method, adding a diagnostic on failure.
func clientFromProviderData(providerData any) (*client.Client, error) {
	if providerData == nil {
		return nil, nil
	}
	c, ok := providerData.(*client.Client)
	if !ok {
		return nil, fmt.Errorf("expected *client.Client, got %T; this is a bug in the provider", providerData)
	}
	return c, nil
}

// optionalString returns a null types.String when v is empty, otherwise a known
// value. Used so optional/computed attributes round-trip cleanly.
func optionalString(v string) types.String {
	if v == "" {
		return types.StringNull()
	}
	return types.StringValue(v)
}
