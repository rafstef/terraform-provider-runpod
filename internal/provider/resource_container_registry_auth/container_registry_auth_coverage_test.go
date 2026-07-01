package resource_container_registry_auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Create error paths ---

// TestCreate_MissingAPIKey asserts Create returns a diagnostics error when
// RUNPOD_API_KEY is unset, before any HTTP call is made.
func TestCreate_MissingAPIKey(t *testing.T) {
	ctx := context.Background()
	sch := ContainerRegistryAuthResourceSchema(ctx)

	m := ContainerRegistryAuthModel{
		Id:       types.StringNull(),
		Name:     types.StringValue("my-registry"),
		Password: types.StringValue("s3cret"),
		Username: types.StringValue("alice"),
	}
	st := tfsdk.State{Schema: sch}
	if d := st.Set(ctx, &m); d.HasError() {
		t.Fatalf("build config state: %v", d)
	}
	cfg := tfsdk.Config{Schema: sch, Raw: st.Raw}

	t.Setenv("RUNPOD_API_KEY", "")

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: sch}}
	(&ContainerRegistryAuthResource{}).Create(ctx, resource.CreateRequest{Config: cfg}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error when RUNPOD_API_KEY is empty, got none")
	}
	if !strings.Contains(resp.Diagnostics.Errors()[0].Detail(), "RUNPOD_API_KEY") {
		t.Errorf("error detail = %q, want mention of RUNPOD_API_KEY", resp.Diagnostics.Errors()[0].Detail())
	}
}

// TestCreate_Non200Status asserts Create returns a diagnostics error when the
// API responds with a non-200 status (parseable JSON body so the status check
// is reached, not the JSON-parse branch).
func TestCreate_Non200Status(t *testing.T) {
	ctx := context.Background()
	sch := ContainerRegistryAuthResourceSchema(ctx)

	m := ContainerRegistryAuthModel{
		Id:       types.StringNull(),
		Name:     types.StringValue("my-registry"),
		Password: types.StringValue("s3cret"),
		Username: types.StringValue("alice"),
	}
	st := tfsdk.State{Schema: sch}
	if d := st.Set(ctx, &m); d.HasError() {
		t.Fatalf("build config state: %v", d)
	}
	cfg := tfsdk.Config{Schema: sch, Raw: st.Raw}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer srv.Close()

	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: sch}}
	(&ContainerRegistryAuthResource{}).Create(ctx, resource.CreateRequest{Config: cfg}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error on non-200 status, got none")
	}
	if !strings.Contains(resp.Diagnostics.Errors()[0].Detail(), "400") {
		t.Errorf("error detail = %q, want status 400 mentioned", resp.Diagnostics.Errors()[0].Detail())
	}
}

// TestCreate_UnparseableJSON asserts Create returns a diagnostics error when the
// 200 response body is not valid JSON.
func TestCreate_UnparseableJSON(t *testing.T) {
	ctx := context.Background()
	sch := ContainerRegistryAuthResourceSchema(ctx)

	m := ContainerRegistryAuthModel{
		Id:       types.StringNull(),
		Name:     types.StringValue("my-registry"),
		Password: types.StringValue("s3cret"),
		Username: types.StringValue("alice"),
	}
	st := tfsdk.State{Schema: sch}
	if d := st.Set(ctx, &m); d.HasError() {
		t.Fatalf("build config state: %v", d)
	}
	cfg := tfsdk.Config{Schema: sch, Raw: st.Raw}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: sch}}
	(&ContainerRegistryAuthResource{}).Create(ctx, resource.CreateRequest{Config: cfg}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error on unparseable JSON, got none")
	}
}

// TestCreate_NoIDInResponse asserts Create returns a diagnostics error when the
// 200 response body parses but omits the "id" field.
func TestCreate_NoIDInResponse(t *testing.T) {
	ctx := context.Background()
	sch := ContainerRegistryAuthResourceSchema(ctx)

	m := ContainerRegistryAuthModel{
		Id:       types.StringNull(),
		Name:     types.StringValue("my-registry"),
		Password: types.StringValue("s3cret"),
		Username: types.StringValue("alice"),
	}
	st := tfsdk.State{Schema: sch}
	if d := st.Set(ctx, &m); d.HasError() {
		t.Fatalf("build config state: %v", d)
	}
	cfg := tfsdk.Config{Schema: sch, Raw: st.Raw}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"my-registry","username":"alice"}`))
	}))
	defer srv.Close()

	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: sch}}
	(&ContainerRegistryAuthResource{}).Create(ctx, resource.CreateRequest{Config: cfg}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error when response omits id, got none")
	}
	if !strings.Contains(resp.Diagnostics.Errors()[0].Detail(), "ID") {
		t.Errorf("error detail = %q, want mention of ID", resp.Diagnostics.Errors()[0].Detail())
	}
}

// --- Read error / not-found / field-mapping paths ---

// TestRead_MissingAPIKey asserts Read returns a diagnostics error when
// RUNPOD_API_KEY is unset.
func TestRead_MissingAPIKey(t *testing.T) {
	ctx := context.Background()
	sch := ContainerRegistryAuthResourceSchema(ctx)

	m := ContainerRegistryAuthModel{
		Id:       types.StringValue("cra-1"),
		Name:     types.StringValue("n"),
		Password: types.StringValue("p"),
		Username: types.StringValue("u"),
	}
	state := tfsdk.State{Schema: sch}
	if d := state.Set(ctx, &m); d.HasError() {
		t.Fatalf("build read state: %v", d)
	}

	t.Setenv("RUNPOD_API_KEY", "")

	resp := &resource.ReadResponse{State: tfsdk.State{Schema: sch}}
	(&ContainerRegistryAuthResource{}).Read(ctx, resource.ReadRequest{State: state}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error when RUNPOD_API_KEY is empty, got none")
	}
}

// TestRead_404_RemovesState asserts CE-1654-style fix for container_registry_auth:
// when a resource is gone (404), Read must call resp.State.RemoveResource so the
// deleted resource is removed from state.
func TestRead_404_RemovesState(t *testing.T) {
	ctx := context.Background()
	sch := ContainerRegistryAuthResourceSchema(ctx)

	m := ContainerRegistryAuthModel{
		Id:       types.StringValue("cra-gone"),
		Name:     types.StringValue("n"),
		Password: types.StringValue("p"),
		Username: types.StringValue("u"),
	}
	state := tfsdk.State{Schema: sch}
	if d := state.Set(ctx, &m); d.HasError() {
		t.Fatalf("build read state: %v", d)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.ReadResponse{State: tfsdk.State{Schema: sch}}
	(&ContainerRegistryAuthResource{}).Read(ctx, resource.ReadRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error for 404 not-found, got: %v", resp.Diagnostics.Errors())
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("state was not removed on 404 - CE-1654 fix should remove resource from state")
	}
}

// TestRead_UnparseableJSON asserts Read returns a diagnostics error when the
// response body is not valid JSON.
func TestRead_UnparseableJSON(t *testing.T) {
	ctx := context.Background()
	sch := ContainerRegistryAuthResourceSchema(ctx)

	m := ContainerRegistryAuthModel{
		Id:       types.StringValue("cra-1"),
		Name:     types.StringValue("n"),
		Password: types.StringValue("p"),
		Username: types.StringValue("u"),
	}
	state := tfsdk.State{Schema: sch}
	if d := state.Set(ctx, &m); d.HasError() {
		t.Fatalf("build read state: %v", d)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.ReadResponse{State: tfsdk.State{Schema: sch}}
	(&ContainerRegistryAuthResource{}).Read(ctx, resource.ReadRequest{State: state}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error on unparseable JSON, got none")
	}
}

// TestRead_EmptyResponse asserts Read returns a diagnostics error when the body
// parses to JSON null (result == nil branch).
func TestRead_EmptyResponse(t *testing.T) {
	ctx := context.Background()
	sch := ContainerRegistryAuthResourceSchema(ctx)

	m := ContainerRegistryAuthModel{
		Id:       types.StringValue("cra-1"),
		Name:     types.StringValue("n"),
		Password: types.StringValue("p"),
		Username: types.StringValue("u"),
	}
	state := tfsdk.State{Schema: sch}
	if d := state.Set(ctx, &m); d.HasError() {
		t.Fatalf("build read state: %v", d)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`null`))
	}))
	defer srv.Close()

	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.ReadResponse{State: tfsdk.State{Schema: sch}}
	(&ContainerRegistryAuthResource{}).Read(ctx, resource.ReadRequest{State: state}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error on null/empty response, got none")
	}
	if !strings.Contains(resp.Diagnostics.Errors()[0].Detail(), "Empty response") {
		t.Errorf("error detail = %q, want 'Empty response'", resp.Diagnostics.Errors()[0].Detail())
	}
}

// TestRead_PartialFields asserts Read tolerates a body that omits name and
// username: those state fields are left unchanged (the field-mapping "ok" guards
// are skipped) and no error is raised.
func TestRead_PartialFields(t *testing.T) {
	ctx := context.Background()
	sch := ContainerRegistryAuthResourceSchema(ctx)

	m := ContainerRegistryAuthModel{
		Id:       types.StringValue("cra-1"),
		Name:     types.StringValue("keep-name"),
		Password: types.StringValue("p"),
		Username: types.StringValue("keep-user"),
	}
	state := tfsdk.State{Schema: sch}
	if d := state.Set(ctx, &m); d.HasError() {
		t.Fatalf("build read state: %v", d)
	}

	// Body parses but contains neither name nor username.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"cra-1"}`))
	}))
	defer srv.Close()

	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.ReadResponse{State: tfsdk.State{Schema: sch}}
	(&ContainerRegistryAuthResource{}).Read(ctx, resource.ReadRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read returned diagnostics: %v", resp.Diagnostics.Errors())
	}

	var out ContainerRegistryAuthModel
	if d := resp.State.Get(ctx, &out); d.HasError() {
		t.Fatalf("read result state: %v", d)
	}
	if out.Name.ValueString() != "keep-name" {
		t.Errorf("state Name = %q, want unchanged keep-name", out.Name.ValueString())
	}
	if out.Username.ValueString() != "keep-user" {
		t.Errorf("state Username = %q, want unchanged keep-user", out.Username.ValueString())
	}
}

// --- Delete error paths ---

// TestDelete_MissingAPIKey asserts Delete returns a diagnostics error when
// RUNPOD_API_KEY is unset.
func TestDelete_MissingAPIKey(t *testing.T) {
	ctx := context.Background()
	sch := ContainerRegistryAuthResourceSchema(ctx)

	m := ContainerRegistryAuthModel{
		Id:       types.StringValue("cra-1"),
		Name:     types.StringValue("n"),
		Password: types.StringValue("p"),
		Username: types.StringValue("u"),
	}
	state := tfsdk.State{Schema: sch}
	if d := state.Set(ctx, &m); d.HasError() {
		t.Fatalf("build delete state: %v", d)
	}

	t.Setenv("RUNPOD_API_KEY", "")

	resp := &resource.DeleteResponse{State: tfsdk.State{Schema: sch}}
	(&ContainerRegistryAuthResource{}).Delete(ctx, resource.DeleteRequest{State: state}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error when RUNPOD_API_KEY is empty, got none")
	}
}

// TestDelete_Non204Status asserts Delete returns a diagnostics error when the
// API responds with a status other than 204.
func TestDelete_Non204Status(t *testing.T) {
	ctx := context.Background()
	sch := ContainerRegistryAuthResourceSchema(ctx)

	m := ContainerRegistryAuthModel{
		Id:       types.StringValue("cra-1"),
		Name:     types.StringValue("n"),
		Password: types.StringValue("p"),
		Username: types.StringValue("u"),
	}
	state := tfsdk.State{Schema: sch}
	if d := state.Set(ctx, &m); d.HasError() {
		t.Fatalf("build delete state: %v", d)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	defer srv.Close()

	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	resp := &resource.DeleteResponse{State: tfsdk.State{Schema: sch}}
	(&ContainerRegistryAuthResource{}).Delete(ctx, resource.DeleteRequest{State: state}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error on non-204 status, got none")
	}
	if !strings.Contains(resp.Diagnostics.Errors()[0].Detail(), "500") {
		t.Errorf("error detail = %q, want status 500 mentioned", resp.Diagnostics.Errors()[0].Detail())
	}
}

// --- Update / Metadata / Schema / constructor smoke ---

// TestUpdate_NotSupported asserts Update always returns an error since the
// resource does not support updates.
func TestUpdate_NotSupported(t *testing.T) {
	ctx := context.Background()
	resp := &resource.UpdateResponse{}
	(&ContainerRegistryAuthResource{}).Update(ctx, resource.UpdateRequest{}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected Update to return a diagnostics error, got none")
	}
}

// TestMetadata asserts the resource type name.
func TestMetadata(t *testing.T) {
	ctx := context.Background()
	resp := &resource.MetadataResponse{}
	(&ContainerRegistryAuthResource{}).Metadata(ctx, resource.MetadataRequest{}, resp)

	if resp.TypeName != "runpod_container_registry_auth" {
		t.Errorf("TypeName = %q, want runpod_container_registry_auth", resp.TypeName)
	}
}

// TestSchema asserts Schema populates the response without diagnostics errors.
func TestSchema(t *testing.T) {
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	(&ContainerRegistryAuthResource{}).Schema(ctx, resource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema returned diagnostics: %v", resp.Diagnostics.Errors())
	}
	if len(resp.Schema.Attributes) == 0 {
		t.Errorf("Schema has no attributes")
	}
}

// TestNewContainerRegistryAuthResource asserts the constructor returns a
// non-nil resource of the expected concrete type.
func TestNewContainerRegistryAuthResource(t *testing.T) {
	r := NewContainerRegistryAuthResource()
	if r == nil {
		t.Fatalf("constructor returned nil")
	}
	if _, ok := r.(*ContainerRegistryAuthResource); !ok {
		t.Errorf("constructor returned %T, want *ContainerRegistryAuthResource", r)
	}
}
