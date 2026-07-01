package datasource_gpu_types

import (
	"os"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/runpod/terraform-provider-runpod/internal/provider/client"
)

func NewGpuTypesDataSource() datasource.DataSource {
	return &GpuTypesDataSource{}
}

type GpuTypesDataSource struct{}

func (d *GpuTypesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "runpod_gpu_types"
}

func (d *GpuTypesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = GpuTypesDataSourceSchema(ctx)
}

func (d *GpuTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	query := `
		query GetGpuTypes {
			gpuTypes {
				id
				displayName
				manufacturer
				cuda_cores
				memory_in_gb
				community_price
				secure_price
				secure_cloud
			}
		}
	`

	variables := map[string]interface{}{}

	apiKey := os.Getenv("RUNPOD_API_KEY")
	clientObj := client.NewRunPodClient(apiKey, client.GetGraphQLEndpoint())
	result, err := clientObj.Query(ctx, query, variables)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	if gpus, ok := result["gpuTypes"].([]interface{}); ok {
		models := make([]GpuTypesModel, len(gpus))
		for i, gpu := range gpus {
			if gpuMap, ok := gpu.(map[string]interface{}); ok {
				var id, displayName, manufacturer string
				var cudaCores, memoryInGb float64
				var communityPrice, securePrice float64
				var secureCloud bool
				
				if v, ok := gpuMap["id"].(string); ok {
					id = v
				} else {
					resp.Diagnostics.AddError("API Error", "Field 'id' is missing or not a string in GPU type response")
					return
				}
				
				if v, ok := gpuMap["displayName"].(string); ok {
					displayName = v
				} else {
					resp.Diagnostics.AddError("API Error", "Field 'displayName' is missing or not a string in GPU type response")
					return
				}
				
				if v, ok := gpuMap["manufacturer"].(string); ok {
					manufacturer = v
				} else {
					resp.Diagnostics.AddError("API Error", "Field 'manufacturer' is missing or not a string in GPU type response")
					return
				}
				
				if v, ok := gpuMap["cuda_cores"].(float64); ok {
					cudaCores = v
				} else {
					resp.Diagnostics.AddError("API Error", "Field 'cuda_cores' is missing or not a float64 in GPU type response")
					return
				}
				
				if v, ok := gpuMap["memory_in_gb"].(float64); ok {
					memoryInGb = v
				} else {
					resp.Diagnostics.AddError("API Error", "Field 'memory_in_gb' is missing or not a float64 in GPU type response")
					return
				}
				
				if v, ok := gpuMap["community_price"].(float64); ok {
					communityPrice = v
				} else {
					resp.Diagnostics.AddError("API Error", "Field 'community_price' is missing or not a float64 in GPU type response")
					return
				}
				
				if v, ok := gpuMap["secure_price"].(float64); ok {
					securePrice = v
				} else {
					resp.Diagnostics.AddError("API Error", "Field 'secure_price' is missing or not a float64 in GPU type response")
					return
				}
				
				if v, ok := gpuMap["secure_cloud"].(bool); ok {
					secureCloud = v
				} else {
					resp.Diagnostics.AddError("API Error", "Field 'secure_cloud' is missing or not a bool in GPU type response")
					return
				}
				
				models[i] = GpuTypesModel{
					Id:             types.StringValue(id),
					DisplayName:    types.StringValue(displayName),
					Manufacturer:   types.StringValue(manufacturer),
					CudaCores:      types.Int64Value(int64(cudaCores)),
					MemoryInGb:     types.Int64Value(int64(memoryInGb)),
					CommunityPrice: types.Float64Value(communityPrice),
					SecurePrice:    types.Float64Value(securePrice),
					SecureCloud:    types.BoolValue(secureCloud),
				}
			}
		}
		parent := GpuTypesDataSourceModel{
			GpuTypes: models,
		}
		diags := resp.State.Set(ctx, &parent)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	} else {
		resp.Diagnostics.AddError("API Error", "Failed to parse GPU types from response")
	}
}
