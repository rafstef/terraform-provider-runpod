package datasource_user

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/runpod/terraform-provider-runpod/internal/provider/client"
)

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct{}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "runpod_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = UserDataSourceSchema(ctx)
}

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	query := `
		query GetUser {
			user {
				id
				pubKey
			}
		}
	`

	variables := map[string]interface{}{}

	apiKey := os.Getenv("RUNPOD_API_KEY")
	runpodClient := client.NewRunPodClient(apiKey, "https://api.runpod.io/graphql")
	result, err := runpodClient.Query(ctx, query, variables)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	if data, ok := result["data"].(map[string]interface{}); ok {
		if user, ok := data["user"].(map[string]interface{}); ok {
			model := UserModel{
				Id:     types.StringValue(user["id"].(string)),
				PubKey: types.StringValue(user["pubKey"].(string)),
			}
			diags := resp.State.Set(ctx, &model)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
		} else {
			resp.Diagnostics.AddError("API Error", "User not found in response")
		}
	} else {
		resp.Diagnostics.AddError("API Error", "Failed to get data from response")
	}
}
