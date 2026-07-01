package datasource_data_centers

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/runpod/terraform-provider-runpod/internal/provider/client"
)

func NewDataCentersDataSource() datasource.DataSource {
	return &DataCentersDataSource{}
}

type DataCentersDataSource struct {
	client *client.RunPodClient
}

func (d *DataCentersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*client.RunPodClient)
	}
}

func (d *DataCentersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "runpod_data_centers"
}

func (d *DataCentersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DataCentersDataSourceSchema(ctx)
}

func (d *DataCentersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	query := `
		query GetDataCenters {
			dataCenters {
				id
				name
				location
				globalNetwork
			}
		}
	`

	variables := map[string]interface{}{}

	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "RunPod client is not configured")
		return
	}
	result, err := d.client.Query(ctx, query, variables)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	if dataCenters, ok := result["dataCenters"].([]interface{}); ok {
		models := make([]DataCentersModel, len(dataCenters))
		for i, dc := range dataCenters {
			if dcMap, ok := dc.(map[string]interface{}); ok {
				var id, name, location string
				var globalNetwork bool
				
				if v, ok := dcMap["id"].(string); ok {
					id = v
				} else {
					resp.Diagnostics.AddError("API Error", "Field 'id' is missing or not a string in data center response")
					return
				}
				
				if v, ok := dcMap["name"].(string); ok {
					name = v
				} else {
					resp.Diagnostics.AddError("API Error", "Field 'name' is missing or not a string in data center response")
					return
				}
				
				if v, ok := dcMap["location"].(string); ok {
					location = v
				} else {
					resp.Diagnostics.AddError("API Error", "Field 'location' is missing or not a string in data center response")
					return
				}
				
				if v, ok := dcMap["globalNetwork"].(bool); ok {
					globalNetwork = v
				} else {
					resp.Diagnostics.AddError("API Error", "Field 'globalNetwork' is missing or not a bool in data center response")
					return
				}
				
				models[i] = DataCentersModel{
					Id:            types.StringValue(id),
					Name:          types.StringValue(name),
					Location:      types.StringValue(location),
					GlobalNetwork: types.BoolValue(globalNetwork),
				}
			}
		}
		parent := DataCentersDataSourceModel{
			DataCenters: models,
		}
		diags := resp.State.Set(ctx, &parent)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	} else {
		resp.Diagnostics.AddError("API Error", "Failed to parse data centers from response")
	}
}
