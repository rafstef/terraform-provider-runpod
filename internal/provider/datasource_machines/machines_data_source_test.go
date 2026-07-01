package datasource_machines

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// CE-1652 (GraphQL double-unwrap) is FIXED by PR #20: client.Query() returns the
// inner "data" map and Read() reads result["machines"] directly. This data
// source is still non-functional because of a SEPARATE, pre-existing bug that
// #20 did not touch: Read builds a []MachinesModel slice and calls State.Set
// against the single-object root schema in machines_data_source_gen.go, which
// the framework rejects (slice/object "Value Conversion Error").
//
// This characterizes the current state: the old double-unwrap error is gone
// (proving the #20 fix reached this data source), but Read still errors in
// State.Set. When the schema/Read shape is fixed (schema becomes a list/nested
// attribute, or Read sets a single object), State.Set will succeed — flip this
// to assert the parsed machines list.
func TestMachinesRead_BlockedBySliceSchemaMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"machines":[{"id":"m-123","name":"node-a","location":"US-CA","listed":true,"gpuType":{"id":"g1","displayName":"NVIDIA RTX 4090"},"gpuTotal":8,"secureCloud":true,"dataCenterId":"US-CA-1"}]}}`))
	}))
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_GRAPHQL_URL", srv.URL)

	ctx := context.Background()
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: MachinesDataSourceSchema(ctx)}}
	(&MachinesDataSource{}).Read(ctx, datasource.ReadRequest{}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("Read now succeeds — the slice/object schema bug looks fixed; flip this to assert the machines list")
	}
	for _, d := range resp.Diagnostics.Errors() {
		if strings.Contains(d.Detail(), "Failed to") {
			t.Fatalf("CE-1652 regression: double-unwrap is back: %v", resp.Diagnostics)
		}
	}
}
