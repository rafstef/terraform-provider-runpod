package datasource_pod

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CE-1652 (GraphQL double-unwrap) is FIXED by PR #20: client.Query() returns the
// inner "data" map and Read() reads result["pod"] directly. This data source is
// still non-functional because of a SEPARATE, pre-existing bug that #20 did not
// touch: Read sets Env to an uninitialized types.List{} (nil element type)
// before calling State.Set, which panics inside the framework regardless of the
// unwrap fix (pod_data_source.go).
//
// This characterizes the current state. The stub supplies every field Read
// dereferences, so the only way to reach State.Set (and the Env panic) is for
// the unwrap to be correct — i.e. the panic proves the #20 fix reached this data
// source. When the Env defect is fixed (e.g. types.ListNull(types.StringType)),
// Read will no longer panic and will populate state — flip this to assert the
// pod fields.
func TestPodDataSourceRead_BlockedByUninitializedEnv(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"pod":{
			"id":"p1","name":"my-pod","status":"RUNNING","desiredStatus":"RUNNING",
			"imageName":"runpod/base:latest","machineId":"m1","machineType":"NVIDIA",
			"gpuTypeId":"NVIDIA A100 80GB","gpuCount":2,"costPerHr":1.89,"memoryInGb":128,
			"volumeInGb":50,"volumeMountPath":"/workspace","volumeKey":"vk1",
			"ports":"8888/http","created_at":"2024-01-01T00:00:00Z","dockerArgs":"",
			"env":[],"templateId":"t1","containerDiskInGb":20
		}}}`))
	}))
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_GRAPHQL_URL", srv.URL)

	ctx := context.Background()
	sch := PodDataSourceSchema(ctx)
	cfgState := tfsdk.State{Schema: sch}
	if d := cfgState.Set(ctx, &PodModel{Id: types.StringValue("p1"), Env: types.ListNull(types.StringType)}); d.HasError() {
		t.Fatalf("building config: %v", d)
	}
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: sch}}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected a panic from the uninitialized Env list in State.Set; " +
				"if Read now completes cleanly the Env defect is fixed — flip this to assert pod fields (name == my-pod)")
		}
	}()
	(&PodDataSource{}).Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: sch, Raw: cfgState.Raw}}, resp)
}
