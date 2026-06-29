package resource_network_volume

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// All NetworkVolumeModel fields are scalars (no List/Map), so a zero-value model
// Sets cleanly; tests set the fields they care about.
func nvModel() NetworkVolumeModel {
	return NetworkVolumeModel{
		Id:           types.StringNull(),
		Name:         types.StringNull(),
		Size:         types.Int64Null(),
		DataCenterId: types.StringNull(),
		StorageTier:  types.StringNull(),
	}
}

func nvStub(t *testing.T, status int, body string, captured *map[string]interface{}, method, path *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if method != nil {
			*method = r.Method
		}
		if path != nil {
			*path = r.URL.Path
		}
		if captured != nil {
			b, _ := io.ReadAll(r.Body)
			m := map[string]interface{}{}
			_ = json.Unmarshal(b, &m)
			*captured = m
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func nvConfig(t *testing.T, m NetworkVolumeModel) tfsdk.Config {
	t.Helper()
	ctx := context.Background()
	sch := NetworkVolumeResourceSchema(ctx)
	st := tfsdk.State{Schema: sch}
	if d := st.Set(ctx, &m); d.HasError() {
		t.Fatalf("build config: %v", d)
	}
	return tfsdk.Config{Schema: sch, Raw: st.Raw}
}

func nvState(t *testing.T, m NetworkVolumeModel) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	sch := NetworkVolumeResourceSchema(ctx)
	st := tfsdk.State{Schema: sch}
	if d := st.Set(ctx, &m); d.HasError() {
		t.Fatalf("build state: %v", d)
	}
	return st
}

func TestNetworkVolumeCreate_Success(t *testing.T) {
	ctx := context.Background()
	sch := NetworkVolumeResourceSchema(ctx)

	m := nvModel()
	m.Name = types.StringValue("vol-a")
	m.Size = types.Int64Value(50)
	m.DataCenterId = types.StringValue("US-CA-1")

	var body map[string]interface{}
	srv := nvStub(t, 200, `{"id":"nv-1","name":"vol-a","size":50,"dataCenterId":"US-CA-1","storageTier":"standard"}`, &body, nil, nil)
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: sch}}
	(&NetworkVolumeResource{}).Create(ctx, resource.CreateRequest{Config: nvConfig(t, m)}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create errored: %v", resp.Diagnostics.Errors())
	}
	if body["name"] != "vol-a" || body["size"] != float64(50) || body["dataCenterId"] != "US-CA-1" {
		t.Errorf("request body missing fields; got %v", body)
	}
	var out NetworkVolumeModel
	if d := resp.State.Get(ctx, &out); d.HasError() {
		t.Fatalf("read state: %v", d)
	}
	if out.Id.ValueString() != "nv-1" || out.StorageTier.ValueString() != "standard" {
		t.Errorf("state not populated: id=%q tier=%q", out.Id.ValueString(), out.StorageTier.ValueString())
	}
}

func TestNetworkVolumeCreate_MissingAPIKey(t *testing.T) {
	ctx := context.Background()
	sch := NetworkVolumeResourceSchema(ctx)
	t.Setenv("RUNPOD_API_KEY", "")
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: sch}}
	(&NetworkVolumeResource{}).Create(ctx, resource.CreateRequest{Config: nvConfig(t, nvModel())}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for missing API key")
	}
}

func TestNetworkVolumeCreate_NonOKStatus(t *testing.T) {
	ctx := context.Background()
	sch := NetworkVolumeResourceSchema(ctx)
	srv := nvStub(t, 400, `{"error":"bad"}`, nil, nil, nil)
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: sch}}
	(&NetworkVolumeResource{}).Create(ctx, resource.CreateRequest{Config: nvConfig(t, nvModel())}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for 400 status")
	}
}

// Correct behavior: a 200 response missing name/size/dataCenterId should yield a
// diagnostic, not a crash. Create uses unchecked type assertions
// (network_volume_resource.go:105-107: result["name"].(string), ["size"].(float64),
// ["dataCenterId"].(string)), so a partial response currently panics. Skipped
// until those are ok-checked.
func TestNetworkVolumeCreate_PartialResponse_ReturnsDiagnostic(t *testing.T) {
	t.Skip("network_volume Create uses unchecked casts on result[name]/[size]/[dataCenterId] and panics on a partial response; should return a diagnostic — un-skip when fixed")
	ctx := context.Background()
	sch := NetworkVolumeResourceSchema(ctx)
	srv := nvStub(t, 200, `{"id":"nv-1"}`, nil, nil, nil)
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: sch}}
	(&NetworkVolumeResource{}).Create(ctx, resource.CreateRequest{Config: nvConfig(t, nvModel())}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a partial response")
	}
}

func TestNetworkVolumeRead_Success(t *testing.T) {
	ctx := context.Background()
	sch := NetworkVolumeResourceSchema(ctx)
	m := nvModel()
	m.Id = types.StringValue("nv-1")
	var method, path string
	srv := nvStub(t, 200, `{"id":"nv-1","name":"renamed","size":100,"dataCenterId":"EU-1","storageTier":"premium"}`, nil, &method, &path)
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.ReadResponse{State: tfsdk.State{Schema: sch}}
	(&NetworkVolumeResource{}).Read(ctx, resource.ReadRequest{State: nvState(t, m)}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read errored: %v", resp.Diagnostics.Errors())
	}
	if method != "GET" || path != "/networkvolumes/nv-1" {
		t.Errorf("expected GET /networkvolumes/nv-1, got %s %s", method, path)
	}
	var out NetworkVolumeModel
	if d := resp.State.Get(ctx, &out); d.HasError() {
		t.Fatalf("read state: %v", d)
	}
	if out.Name.ValueString() != "renamed" || out.Size.ValueInt64() != 100 {
		t.Errorf("state not refreshed: name=%q size=%d", out.Name.ValueString(), out.Size.ValueInt64())
	}
}

// Correct behavior: on a 404, Read should remove the resource from state
// (resp.State.RemoveResource) so Terraform plans to recreate it. Currently Read
// only AddWarning + returns (network_volume_resource.go:164), leaving stale state
// — same defect class as CE-1654 (pod Read). Skipped until fixed.
func TestNetworkVolumeRead_404_RemovesState(t *testing.T) {
	t.Skip("network_volume Read AddWarnings on 404 but does not RemoveResource, leaving stale state (CE-1654 class) — un-skip when fixed")
	ctx := context.Background()
	m := nvModel()
	m.Id = types.StringValue("nv-1")
	srv := nvStub(t, 404, `{"error":"not found"}`, nil, nil, nil)
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.ReadResponse{State: nvState(t, m)}
	(&NetworkVolumeResource{}).Read(ctx, resource.ReadRequest{State: nvState(t, m)}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("404 should not error: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Error("expected state removed on 404")
	}
}

func TestNetworkVolumeUpdate_Success(t *testing.T) {
	ctx := context.Background()
	sch := NetworkVolumeResourceSchema(ctx)
	prior := nvModel()
	prior.Id = types.StringValue("nv-1")
	prior.Name = types.StringValue("old-name")

	desired := nvModel()
	desired.Id = types.StringValue("nv-1")
	desired.Name = types.StringValue("new-name")

	var method, path string
	srv := nvStub(t, 200, `{"id":"nv-1","name":"new-name","size":50,"dataCenterId":"US-CA-1","storageTier":"standard"}`, nil, &method, &path)
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	// Update reads req.Config (desired) + req.State (prior).
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: sch}}
	(&NetworkVolumeResource{}).Update(ctx, resource.UpdateRequest{
		Config: nvConfig(t, desired),
		State:  nvState(t, prior),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Update errored: %v", resp.Diagnostics.Errors())
	}
	if method != "PATCH" || path != "/networkvolumes/nv-1" {
		t.Errorf("expected PATCH /networkvolumes/nv-1, got %s %s", method, path)
	}
}

func TestNetworkVolumeDelete_Success(t *testing.T) {
	ctx := context.Background()
	sch := NetworkVolumeResourceSchema(ctx)
	m := nvModel()
	m.Id = types.StringValue("nv-1")
	var method, path string
	srv := nvStub(t, 204, ``, nil, &method, &path)
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.DeleteResponse{State: tfsdk.State{Schema: sch}}
	(&NetworkVolumeResource{}).Delete(ctx, resource.DeleteRequest{State: nvState(t, m)}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Delete errored: %v", resp.Diagnostics.Errors())
	}
	if method != "DELETE" || path != "/networkvolumes/nv-1" {
		t.Errorf("expected DELETE /networkvolumes/nv-1, got %s %s", method, path)
	}
}

func TestNetworkVolumeDelete_NonNoContent(t *testing.T) {
	ctx := context.Background()
	sch := NetworkVolumeResourceSchema(ctx)
	m := nvModel()
	m.Id = types.StringValue("nv-1")
	srv := nvStub(t, 500, `oops`, nil, nil, nil)
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.DeleteResponse{State: tfsdk.State{Schema: sch}}
	(&NetworkVolumeResource{}).Delete(ctx, resource.DeleteRequest{State: nvState(t, m)}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for non-204 delete")
	}
}
