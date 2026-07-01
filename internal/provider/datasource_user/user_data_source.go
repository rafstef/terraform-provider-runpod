package datasource_user

import (
	"os"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/runpod/terraform-provider-runpod/internal/provider/client"
)

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	client *client.RunPodClient
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*client.RunPodClient)
	}
}

func (d *UserDataSource) getClient() *client.RunPodClient {
	if d.client != nil {
		return d.client
	}
	apiKey := os.Getenv("RUNPOD_API_KEY")
	endpoint := os.Getenv("RUNPOD_GRAPHQL_URL")
	if endpoint == "" {
		endpoint = "https://api.runpod.io/graphql"
	}
	baseURL := os.Getenv("RUNPOD_BASE_URL")
	if baseURL == "" {
		baseURL = "https://rest.runpod.io/v1"
	}
	d.client = client.NewRunPodClient(apiKey, endpoint, baseURL)
	return d.client
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "runpod_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = UserDataSourceSchema(ctx)
}

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	query := `
		query GetUser {
			myself {
				id
				pubKey
			}
		}
	`

	variables := map[string]interface{}{}

	client := d.getClient()
	result, err := client.Query(ctx, query, variables)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	if user, ok := result["myself"].(map[string]interface{}); ok {
		var idVal, pubKeyVal string
		
		if id, ok := user["id"].(string); ok {
			idVal = id
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'id' is missing or not a string in user response")
			return
		}
		
		if pubKey, ok := user["pubKey"].(string); ok {
			pubKeyVal = pubKey
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'pubKey' is missing or not a string in user response")
			return
		}
		
		model := UserModel{
			Id:     types.StringValue(idVal),
			PubKey: types.StringValue(pubKeyVal),
		}
		diags := resp.State.Set(ctx, &model)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	} else {
		resp.Diagnostics.AddError("API Error", "User not found in response")
	}
}
