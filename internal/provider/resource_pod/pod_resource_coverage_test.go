package resource_pod

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestPodResource_MetadataAndSchema covers the boilerplate accessors (Metadata,
// Schema, NewPodResource) and guards against a schema-build panic.
func TestPodResource_MetadataAndSchema(t *testing.T) {
	ctx := context.Background()
	r := NewPodResource()
	mResp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "runpod"}, mResp)
	if mResp.TypeName == "" {
		t.Error("Metadata produced an empty TypeName")
	}
	sResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, sResp)
	if len(sResp.Schema.Attributes) == 0 {
		t.Error("Schema produced no attributes")
	}
}

// TestPodRead_FullFieldMapping exercises every field-mapping branch in Read by
// returning a fully-populated body (cloudType, containerDiskInGb, costPerHr,
// created_at, dockerEntrypoint, dockerStartCmd, gpuTypeId, interruptible,
// machineId, memoryInGb, networkVolume, status, templateId, volumeEncrypted,
// volumeInGb).
func TestPodRead_FullFieldMapping(t *testing.T) {
	ctx := context.Background()
	sch := PodResourceSchema(ctx)
	body := `{
		"id":"pod-1","status":"RUNNING","gpuTypeId":"NVIDIA A100","machineId":"m-1",
		"costPerHr":1.5,"created_at":"2024-01-01T00:00:00Z","memoryInGb":16,
		"volumeInGb":50,"containerDiskInGb":20,"templateId":"t-1","cloudType":"SECURE",
		"networkVolume":{"id":"nv-1"},"dockerEntrypoint":["/bin/sh"],
		"dockerStartCmd":["run"],"interruptible":true,"volumeEncrypted":true
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	m := baseModel()
	m.Id = types.StringValue("pod-1")
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: sch}}
	(&PodResource{}).Read(ctx, resource.ReadRequest{State: podState(t, m)}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Read errored: %v", resp.Diagnostics.Errors())
	}

	var out PodModel
	if d := resp.State.Get(ctx, &out); d.HasError() {
		t.Fatalf("read state: %v", d)
	}
	if out.MachineId.ValueString() != "m-1" {
		t.Errorf("MachineId = %q, want m-1", out.MachineId.ValueString())
	}
	if out.CostPerHr.ValueFloat64() != 1.5 {
		t.Errorf("CostPerHr = %v, want 1.5", out.CostPerHr.ValueFloat64())
	}
	if out.ContainerDiskInGb.ValueInt64() != 20 {
		t.Errorf("ContainerDiskInGb = %d, want 20", out.ContainerDiskInGb.ValueInt64())
	}
	if !out.Interruptible.ValueBool() || !out.VolumeEncrypted.ValueBool() {
		t.Errorf("interruptible/volumeEncrypted not mapped: %v / %v", out.Interruptible, out.VolumeEncrypted)
	}
	if out.NetworkVolumeId.ValueString() != "nv-1" {
		t.Errorf("NetworkVolumeId = %q, want nv-1 (from networkVolume.id)", out.NetworkVolumeId.ValueString())
	}
	if out.DockerEntrypoint.IsNull() || len(out.DockerEntrypoint.Elements()) != 1 {
		t.Errorf("DockerEntrypoint not mapped: %v", out.DockerEntrypoint)
	}
}

// TestPodUpdate_ManyFieldsInBody exercises Update's conditional body-build
// branches by changing many fields vs prior state and asserting the PATCH body.
func TestPodUpdate_ManyFieldsInBody(t *testing.T) {
	ctx := context.Background()
	sch := PodResourceSchema(ctx)

	prior := baseModel()
	prior.Id = types.StringValue("pod-1")
	prior.Name = types.StringValue("old")
	prior.GpuTypeId = types.StringValue("NVIDIA GPU")
	prior.MachineId = types.StringValue("m-old")
	prior.ContainerDiskInGb = types.Int64Value(20)
	prior.VolumeKey = types.StringValue("oldkey")

	desired := baseModel()
	desired.Id = types.StringValue("pod-1")
	desired.Name = types.StringValue("new")
	desired.GpuTypeId = types.StringValue("NVIDIA A100")
	desired.MachineId = types.StringValue("m-1")
	desired.ContainerDiskInGb = types.Int64Value(30)
	desired.VolumeKey = types.StringValue("mykey")
	desired.GpuCount = types.Int64Value(2)
	desired.CloudType = types.StringValue("SECURE")
	desired.DockerArgs = types.StringValue("--foo")
	desired.Env = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("K=V")})
	desired.Port = types.Int64Value(8080)
	desired.Ports = types.StringValue("8080/http")
	desired.StartSsh = types.BoolValue(true)
	desired.StartJupyter = types.BoolValue(true)
	desired.StopAfter = types.StringValue("1h")
	desired.TerminateAfter = types.StringValue("2h")
	desired.VolumeInGb = types.Float64Value(50)
	desired.VolumeMountPath = types.StringValue("/data")
	desired.BidPerGpu = types.Float64Value(1.5)

	var body map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		_, _ = w.Write([]byte(`{"id":"pod-1"}`))
	}))
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: sch}}
	(&PodResource{}).Update(ctx, resource.UpdateRequest{Config: podConfig(t, desired), State: podState(t, prior)}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Update errored: %v", resp.Diagnostics.Errors())
	}

	for _, k := range []string{"name", "gpuCount", "cloudType", "dockerArgs", "env", "startSsh", "startJupyter", "stopAfter", "terminateAfter", "volumeMountPath", "gpuTypeId", "machineId", "containerDiskInGb", "volumeKey", "port", "bidPerGpu"} {
		if _, ok := body[k]; !ok {
			t.Errorf("PATCH body missing %q; got %v", k, body)
		}
	}
	if body["name"] != "new" {
		t.Errorf("body name = %v, want new", body["name"])
	}
}
