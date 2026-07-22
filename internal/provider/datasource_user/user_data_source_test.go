package datasource_user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// TestUserRead_PopulatesState tests the REST migration with v2 user endpoint.
func TestUserRead_PopulatesState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// v2 REST response format with data envelope
		_, _ = w.Write([]byte(`{
			"data": {
				"id": "u1",
				"name": "Test User",
				"email": "test@runpod.io",
				"pubKey": "ssh-ed25519 AAAA test@runpod",
				"verified": true,
				"cloudType": "aws",
				"gpuLimit": 10,
				"gpuUsage": 2,
				"storageLimit": 100,
				"storageUsage": 25
			},
			"meta": {},
			"error": null
		}`))
	}))
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	ctx := context.Background()
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: UserDataSourceSchema(ctx)}}
	(&UserDataSource{}).Read(ctx, datasource.ReadRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error, got: %v", resp.Diagnostics)
	}

	var model UserModel
	diags := resp.State.Get(ctx, &model)
	if diags.HasError() {
		t.Fatalf("failed to read state: %v", diags)
	}
	if model.Id.ValueString() != "u1" {
		t.Errorf("expected id %q, got %q", "u1", model.Id.ValueString())
	}
	if model.Name.ValueString() != "Test User" {
		t.Errorf("expected name %q, got %q", "Test User", model.Name.ValueString())
	}
	if model.Email.ValueString() != "test@runpod.io" {
		t.Errorf("expected email %q, got %q", "test@runpod.io", model.Email.ValueString())
	}
	if model.PubKey.ValueString() != "ssh-ed25519 AAAA test@runpod" {
		t.Errorf("expected pubKey %q, got %q", "ssh-ed25519 AAAA test@runpod", model.PubKey.ValueString())
	}
	if !model.Verified.ValueBool() {
		t.Errorf("expected verified to be true")
	}
	if model.CloudType.ValueString() != "aws" {
		t.Errorf("expected cloudType %q, got %q", "aws", model.CloudType.ValueString())
	}
	if model.GpuLimit.ValueFloat64() != 10 {
		t.Errorf("expected gpuLimit %v, got %v", 10.0, model.GpuLimit.ValueFloat64())
	}
	if model.GpuUsage.ValueFloat64() != 2 {
		t.Errorf("expected gpuUsage %v, got %v", 2.0, model.GpuUsage.ValueFloat64())
	}
	if model.StorageLimit.ValueFloat64() != 100 {
		t.Errorf("expected storageLimit %v, got %v", 100.0, model.StorageLimit.ValueFloat64())
	}
	if model.StorageUsage.ValueFloat64() != 25 {
		t.Errorf("expected storageUsage %v, got %v", 25.0, model.StorageUsage.ValueFloat64())
	}
}

// TestUserRead_HandlesMissingOptionalFields tests that optional fields can be omitted.
func TestUserRead_HandlesMissingOptionalFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"data": {
				"id": "u2",
				"name": "Minimal User"
			}
		}`))
	}))
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	ctx := context.Background()
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: UserDataSourceSchema(ctx)}}
	(&UserDataSource{}).Read(ctx, datasource.ReadRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error for minimal response, got: %v", resp.Diagnostics)
	}

	var model UserModel
	diags := resp.State.Get(ctx, &model)
	if diags.HasError() {
		t.Fatalf("failed to read state: %v", diags)
	}
	if model.Id.ValueString() != "u2" {
		t.Errorf("expected id %q, got %q", "u2", model.Id.ValueString())
	}
	if model.Name.ValueString() != "Minimal User" {
		t.Errorf("expected name %q, got %q", "Minimal User", model.Name.ValueString())
	}
}

// TestUserRead_MissingIdField_AddsDiagnostic tests that missing id field causes an error.
func TestUserRead_MissingIdField_AddsDiagnostic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data": {"name": "No ID User"}}`))
	}))
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	ctx := context.Background()
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: UserDataSourceSchema(ctx)}}
	(&UserDataSource{}).Read(ctx, datasource.ReadRequest{}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics error when id is missing")
	}
	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == "API Error" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected an \"API Error\" diagnostic, got: %v", resp.Diagnostics)
	}
}

// TestUserRead_RestError_AddsDiagnostic tests REST API error handling.
func TestUserRead_RestError_AddsDiagnostic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"Internal Server Error"}`))
	}))
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_BASE_URL", srv.URL)

	ctx := context.Background()
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: UserDataSourceSchema(ctx)}}
	(&UserDataSource{}).Read(ctx, datasource.ReadRequest{}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics error from non-200 HTTP response")
	}
}
