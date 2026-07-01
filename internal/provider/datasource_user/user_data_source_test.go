package datasource_user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Positive regression for CE-1652 (fixed by PR #20).
// After Query strips the outer {"data":...} envelope, Read reads result["user"]
// and dereferences user["id"] and user["pubKey"] into state. This asserts the
// fixed behavior: no diagnostics error and state is populated from the response.
func TestUserRead_PopulatesState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"myself":{"id":"u1","pubKey":"ssh-ed25519 AAAA test@runpod"}}}`))
	}))
	defer srv.Close()
	t.Setenv("RUNPOD_API_KEY", "testkey123")
	t.Setenv("RUNPOD_GRAPHQL_URL", srv.URL)

	ctx := context.Background()
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: UserDataSourceSchema(ctx)}}
	(&UserDataSource{}).Read(ctx, datasource.ReadRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error after CE-1652 fix, got: %v", resp.Diagnostics)
	}

	var model UserModel
	diags := resp.State.Get(ctx, &model)
	if diags.HasError() {
		t.Fatalf("failed to read state: %v", diags)
	}
	if model.Id.ValueString() != "u1" {
		t.Errorf("expected id %q, got %q", "u1", model.Id.ValueString())
	}
	if model.PubKey.ValueString() != "ssh-ed25519 AAAA test@runpod" {
		t.Errorf("expected pubKey %q, got %q", "ssh-ed25519 AAAA test@runpod", model.PubKey.ValueString())
	}
}
