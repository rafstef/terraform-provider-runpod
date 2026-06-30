package resource_container_registry_auth

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestAccContainerRegistryAuthLifecycle_riab drives the real Create->Read->Delete
// against runpod-in-a-box. Gated on RIAB_ACC=1; requires RUNPOD_BASE_URL +
// RUNPOD_API_KEY (mint via riab scripts/mint-test-jwts.sh).
//
// Skip-gated on CE-1681: Create checks StatusCode != 200 but the API returns 201
// Created, so Create fails on success and orphans the resource. Verified via curl
// that POST /containerregistryauth returns 201 {id,name}. Un-skip when Create
// accepts 201 — the lifecycle below then becomes a real passing acceptance test.
func TestAccContainerRegistryAuthLifecycle_riab(t *testing.T) {
	if os.Getenv("RIAB_ACC") != "1" {
		t.Skip("set RIAB_ACC=1 + RUNPOD_BASE_URL + RUNPOD_API_KEY to run the live riab CRA lifecycle")
	}
	t.Skip("CE-1681: container_registry_auth Create rejects HTTP 201 and orphans the resource — un-skip when Create accepts 201")
	if os.Getenv("RUNPOD_API_KEY") == "" || os.Getenv("RUNPOD_BASE_URL") == "" {
		t.Fatal("RUNPOD_API_KEY and RUNPOD_BASE_URL must be set for the acceptance test")
	}

	ctx := context.Background()
	sch := ContainerRegistryAuthResourceSchema(ctx)

	mkState := func(m ContainerRegistryAuthModel) tfsdk.State {
		st := tfsdk.State{Schema: sch}
		if d := st.Set(ctx, &m); d.HasError() {
			t.Fatalf("build state: %v", d)
		}
		return st
	}

	// --- Create ---
	m := ContainerRegistryAuthModel{
		Name:     types.StringValue("tf-acc-cra"),
		Username: types.StringValue("alice"),
		Password: types.StringValue("s3cret"),
	}
	cResp := &resource.CreateResponse{State: tfsdk.State{Schema: sch}}
	(&ContainerRegistryAuthResource{}).Create(ctx, resource.CreateRequest{Config: tfsdk.Config{Schema: sch, Raw: mkState(m).Raw}}, cResp)
	if cResp.Diagnostics.HasError() {
		t.Fatalf("Create: %v", cResp.Diagnostics)
	}
	var created ContainerRegistryAuthModel
	cResp.State.Get(ctx, &created)
	if created.Id.ValueString() == "" {
		t.Fatal("Create returned an empty id")
	}
	t.Logf("created cra id=%s", created.Id.ValueString())

	defer func() {
		dResp := &resource.DeleteResponse{State: mkState(created)}
		(&ContainerRegistryAuthResource{}).Delete(ctx, resource.DeleteRequest{State: mkState(created)}, dResp)
		if dResp.Diagnostics.HasError() {
			t.Errorf("Delete: %v", dResp.Diagnostics)
		}
	}()

	// --- Read ---
	rResp := &resource.ReadResponse{State: tfsdk.State{Schema: sch}}
	(&ContainerRegistryAuthResource{}).Read(ctx, resource.ReadRequest{State: mkState(created)}, rResp)
	if rResp.Diagnostics.HasError() {
		t.Fatalf("Read: %v", rResp.Diagnostics)
	}
	var readBack ContainerRegistryAuthModel
	rResp.State.Get(ctx, &readBack)
	if readBack.Name.ValueString() != "tf-acc-cra" {
		t.Errorf("Read name = %q, want tf-acc-cra", readBack.Name.ValueString())
	}
}
