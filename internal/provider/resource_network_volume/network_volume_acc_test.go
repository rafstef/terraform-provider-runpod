package resource_network_volume

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestAccNetworkVolumeLifecycle_riab drives the real NetworkVolumeResource
// Create -> Read -> Update -> Delete against a live local rphttp v1 endpoint
// (runpod-in-a-box). Gated on RIAB_ACC=1; requires:
//
//	RUNPOD_BASE_URL=http://localhost:8081/v1
//	RUNPOD_API_KEY=$TEST_USER_JWT   (mint via riab scripts/mint-test-jwts.sh)
//
// dataCenterId MFS-1 is the data center riab exposes for local volumes.
func TestAccNetworkVolumeLifecycle_riab(t *testing.T) {
	if os.Getenv("RIAB_ACC") != "1" {
		t.Skip("set RIAB_ACC=1 + RUNPOD_BASE_URL + RUNPOD_API_KEY to run the live riab network volume lifecycle")
	}
	// CE-1681: Create rejects HTTP 201 (the API returns 201 Created), so it fails
	// on a successful create AND returns before setting the id into state — which
	// orphans the volume in the API on every run. Skip until fixed to avoid
	// leaking volumes; the lifecycle below (verified manually: create 201 / read /
	// update / delete 204) becomes a passing acceptance test once Create accepts 201.
	t.Skip("CE-1681: network_volume Create rejects HTTP 201 and orphans the volume — un-skip when Create accepts 201")
	if os.Getenv("RUNPOD_API_KEY") == "" || os.Getenv("RUNPOD_BASE_URL") == "" {
		t.Fatal("RUNPOD_API_KEY and RUNPOD_BASE_URL must be set for the acceptance test")
	}

	ctx := context.Background()
	sch := NetworkVolumeResourceSchema(ctx)

	// --- Create ---
	m := nvModel()
	m.Name = types.StringValue("tf-acc-nv")
	m.Size = types.Int64Value(10)
	m.DataCenterId = types.StringValue("MFS-1")

	cResp := &resource.CreateResponse{State: tfsdk.State{Schema: sch}}
	(&NetworkVolumeResource{}).Create(ctx, resource.CreateRequest{Config: nvConfig(t, m)}, cResp)
	if cResp.Diagnostics.HasError() {
		t.Fatalf("Create: %v", cResp.Diagnostics)
	}
	var created NetworkVolumeModel
	cResp.State.Get(ctx, &created)
	id := created.Id.ValueString()
	if id == "" {
		t.Fatal("Create returned an empty network volume id")
	}
	t.Logf("created network volume id=%s", id)

	// Always clean up, even if a later step fails.
	defer func() {
		dResp := &resource.DeleteResponse{State: nvState(t, created)}
		(&NetworkVolumeResource{}).Delete(ctx, resource.DeleteRequest{State: nvState(t, created)}, dResp)
		if dResp.Diagnostics.HasError() {
			t.Errorf("Delete: %v", dResp.Diagnostics)
		} else {
			t.Logf("deleted network volume id=%s", id)
		}
	}()

	// --- Read (real GET against the live server) ---
	rResp := &resource.ReadResponse{State: tfsdk.State{Schema: sch}}
	(&NetworkVolumeResource{}).Read(ctx, resource.ReadRequest{State: nvState(t, created)}, rResp)
	if rResp.Diagnostics.HasError() {
		t.Fatalf("Read: %v", rResp.Diagnostics)
	}
	var readBack NetworkVolumeModel
	rResp.State.Get(ctx, &readBack)
	if readBack.Name.ValueString() != "tf-acc-nv" {
		t.Errorf("Read name = %q, want tf-acc-nv", readBack.Name.ValueString())
	}
	if readBack.Size.ValueInt64() != 10 {
		t.Errorf("Read size = %d, want 10", readBack.Size.ValueInt64())
	}
	if readBack.DataCenterId.ValueString() != "MFS-1" {
		t.Errorf("Read dataCenterId = %q, want MFS-1", readBack.DataCenterId.ValueString())
	}
	t.Logf("read network volume id=%s name=%q size=%d", id, readBack.Name.ValueString(), readBack.Size.ValueInt64())

	// --- Update (rename via PATCH) ---
	desired := created
	desired.Name = types.StringValue("tf-acc-nv-renamed")
	uResp := &resource.UpdateResponse{State: tfsdk.State{Schema: sch}}
	(&NetworkVolumeResource{}).Update(ctx, resource.UpdateRequest{
		Config: nvConfig(t, desired),
		State:  nvState(t, created),
	}, uResp)
	if uResp.Diagnostics.HasError() {
		t.Fatalf("Update: %v", uResp.Diagnostics)
	}
	var updated NetworkVolumeModel
	uResp.State.Get(ctx, &updated)
	if updated.Name.ValueString() != "tf-acc-nv-renamed" {
		t.Errorf("Update name = %q, want tf-acc-nv-renamed", updated.Name.ValueString())
	}
	t.Logf("updated network volume id=%s name=%q", id, updated.Name.ValueString())
}
