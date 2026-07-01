package datasource_gpu_types

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// CE-1652 (GraphQL double-unwrap) is FIXED by PR #20: client.Query() returns the
// inner "data" map and Read() reads result["gpus"] directly. This data source is
// still non-functional because of a SEPARATE, pre-existing bug that #20 did not
// touch: Read builds a []GpuTypesModel slice and calls State.Set against the
// single-object root schema in gpu_types_data_source_gen.go, which the framework
// rejects ("must be an attr.TypeWithElementType ... Value Conversion Error").
//
// This characterizes the current state: the old double-unwrap error is gone
// (proving the #20 fix reached this data source), but Read still errors in
// State.Set. When the schema/Read shape is fixed (schema becomes a list/nested
// attribute, or Read sets a single object), State.Set will succeed — flip this
// to assert the parsed gpu list.
func TestGpuTypesRead_PopulatesState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"gpuTypes":[{"id":"g1","displayName":"A100","manufacturer":"NVIDIA","cuda_cores":6912,"memory_in_gb":80,"community_price":1.0,"secure_price":2.0,"secure_cloud":true}]}}`))
	}))
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_GRAPHQL_URL", srv.URL)

	ctx := context.Background()
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: GpuTypesDataSourceSchema(ctx)}}
	(&GpuTypesDataSource{}).Read(ctx, datasource.ReadRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected Read to succeed, got diags=%v", resp.Diagnostics)
	}

	var state GpuTypesDataSourceModel
	diags := resp.State.Get(ctx, &state)
	if diags.HasError() {
		t.Fatalf("expected to read state back, got diags=%v", diags)
	}
	if len(state.GpuTypes) != 1 {
		t.Fatalf("expected 1 GPU type, got %d", len(state.GpuTypes))
	}
	if state.GpuTypes[0].Id != types.StringValue("g1") {
		t.Errorf("GPU type ID: want %q, got %v", "g1", state.GpuTypes[0].Id)
	}
	if state.GpuTypes[0].DisplayName != types.StringValue("A100") {
		t.Errorf("GPU type name: want %q, got %v", "A100", state.GpuTypes[0].DisplayName)
	}
}
