package datasource_container_registry_auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// CE-1671 (double-unwrap) is FIXED: client.Query() returns the inner "data" map
// and Read() reads result["containerRegistryAuths"] directly. This data source
// is still non-functional because of a SEPARATE, pre-existing bug: Read builds
// a []ContainerRegistryAuthModel slice and calls State.Set against the
// single-object root schema (id/name/username flat), which the framework
// rejects ("must be an attr.TypeWithElementType ... Value Conversion Error").
//
// This characterizes the current state: the old double-unwrap error is gone
// (proving the fix reached this data source), but Read still errors in
// State.Set. When the schema/Read shape is fixed (schema becomes a list/nested
// attribute, or Read sets a single object), State.Set will succeed.
func TestContainerRegistryAuthDataSourceRead_BlockedBySliceSchemaMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"containerRegistryAuths":[
			{"id":"auth-1","name":"dockerhub","username":"alice"},
			{"id":"auth-2","name":"ghcr","username":"bob"}
		]}}`))
	}))
	defer srv.Close()

	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_GRAPHQL_URL", srv.URL)

	ctx := context.Background()
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: ContainerRegistryAuthDataSourceSchema(ctx)}}
	(&ContainerRegistryAuthDataSource{}).Read(ctx, datasource.ReadRequest{}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("Read now succeeds — the slice/object schema bug looks fixed; flip this to assert the auth list")
	}
	// The CE-1671 unwrap-failure branch reported "Failed to ... from response".
	// After the fix that must be gone; the only remaining error is the unrelated
	// State.Set slice/object conversion.
	for _, d := range resp.Diagnostics.Errors() {
		if strings.Contains(d.Detail(), "Failed to") {
			t.Fatalf("CE-1671 regression: double-unwrap is back: %v", resp.Diagnostics)
		}
	}
}
