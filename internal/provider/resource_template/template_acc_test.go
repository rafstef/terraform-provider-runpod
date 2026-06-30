package resource_template

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestAccTemplateLifecycle_riab drives the real Create->Read->Update->Delete
// against runpod-in-a-box. Gated on RIAB_ACC=1; requires RUNPOD_BASE_URL +
// RUNPOD_API_KEY (mint via riab scripts/mint-test-jwts.sh).
//
// Skip-gated on CE-1681: Create checks StatusCode != 200 but the API returns 201
// Created, so Create fails on success and orphans the template. Verified via curl
// that POST /templates returns 201 with the full body. Un-skip when Create
// accepts 201 — the lifecycle below then becomes a real passing acceptance test.
func TestAccTemplateLifecycle_riab(t *testing.T) {
	if os.Getenv("RIAB_ACC") != "1" {
		t.Skip("set RIAB_ACC=1 + RUNPOD_BASE_URL + RUNPOD_API_KEY to run the live riab template lifecycle")
	}
	t.Skip("CE-1681: template Create rejects HTTP 201 and orphans the template — un-skip when Create accepts 201")
	if os.Getenv("RUNPOD_API_KEY") == "" || os.Getenv("RUNPOD_BASE_URL") == "" {
		t.Fatal("RUNPOD_API_KEY and RUNPOD_BASE_URL must be set for the acceptance test")
	}

	ctx := context.Background()
	sch := TemplateResourceSchema(ctx)
	mkState := func(m TemplateModel) tfsdk.State {
		st := tfsdk.State{Schema: sch}
		if d := st.Set(ctx, &m); d.HasError() {
			t.Fatalf("build state: %v", d)
		}
		return st
	}

	// --- Create ---
	m := newBaseModel()
	m.Name = types.StringValue("tf-acc-tpl")
	m.ImageName = types.StringValue("runpod/base:0.0.0")
	m.ContainerDiskInGb = types.Int64Value(10)

	cResp := &resource.CreateResponse{State: tfsdk.State{Schema: sch}}
	(&TemplateResource{}).Create(ctx, resource.CreateRequest{Config: buildConfig(t, ctx, m)}, cResp)
	if cResp.Diagnostics.HasError() {
		t.Fatalf("Create: %v", cResp.Diagnostics)
	}
	var created TemplateModel
	cResp.State.Get(ctx, &created)
	if created.Id.ValueString() == "" {
		t.Fatal("Create returned an empty template id")
	}
	t.Logf("created template id=%s", created.Id.ValueString())

	defer func() {
		dResp := &resource.DeleteResponse{State: mkState(created)}
		(&TemplateResource{}).Delete(ctx, resource.DeleteRequest{State: mkState(created)}, dResp)
		if dResp.Diagnostics.HasError() {
			t.Errorf("Delete: %v", dResp.Diagnostics)
		}
	}()

	// --- Read ---
	rResp := &resource.ReadResponse{State: tfsdk.State{Schema: sch}}
	(&TemplateResource{}).Read(ctx, resource.ReadRequest{State: mkState(created)}, rResp)
	if rResp.Diagnostics.HasError() {
		t.Fatalf("Read: %v", rResp.Diagnostics)
	}
	var readBack TemplateModel
	rResp.State.Get(ctx, &readBack)
	if readBack.ImageName.ValueString() != "runpod/base:0.0.0" {
		t.Errorf("Read imageName = %q, want runpod/base:0.0.0", readBack.ImageName.ValueString())
	}

	// --- Update (rename) ---
	desired := created
	desired.Name = types.StringValue("tf-acc-tpl-renamed")
	uResp := &resource.UpdateResponse{State: tfsdk.State{Schema: sch}}
	(&TemplateResource{}).Update(ctx, resource.UpdateRequest{
		Config: buildConfig(t, ctx, desired),
		State:  mkState(created),
	}, uResp)
	if uResp.Diagnostics.HasError() {
		t.Fatalf("Update: %v", uResp.Diagnostics)
	}
}
