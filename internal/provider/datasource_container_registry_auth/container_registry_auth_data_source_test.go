package datasource_container_registry_auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// CE-1671 (double-unwrap) is FIXED: client.Query() returns the inner "data" map
// and Read() reads result["containerRegistryAuths"] directly. CE-1675 (schema
// shape) is FIXED: schema now uses a ListNestedAttribute and Read sets the parent
// model with the populated list. Read now succeeds and populates state.
func TestContainerRegistryAuthDataSourceRead_PopulatesState(t *testing.T) {
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

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected Read to succeed, got diags=%v", resp.Diagnostics)
	}

	var state ContainerRegistryAuthDataSourceModel
	diags := resp.State.Get(ctx, &state)
	if diags.HasError() {
		t.Fatalf("expected to read state back, got diags=%v", diags)
	}
	if len(state.ContainerRegistryAuths) != 2 {
		t.Fatalf("expected 2 auths, got %d", len(state.ContainerRegistryAuths))
	}
	if state.ContainerRegistryAuths[0].Name != types.StringValue("dockerhub") {
		t.Errorf("first auth name: want %q, got %v", "dockerhub", state.ContainerRegistryAuths[0].Name)
	}
	if state.ContainerRegistryAuths[1].Name != types.StringValue("ghcr") {
		t.Errorf("second auth name: want %q, got %v", "ghcr", state.ContainerRegistryAuths[1].Name)
	}
}
