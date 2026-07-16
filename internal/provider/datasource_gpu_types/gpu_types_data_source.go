package datasource_gpu_types

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	client "github.com/runpod/terraform-provider-runpod/internal/provider/client"
)

func NewGpuTypesDataSource() datasource.DataSource {
	return &GpuTypesDataSource{}
}

type GpuTypesDataSource struct {
	client *client.RunPodClient
}

func (d *GpuTypesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*client.RunPodClient)
	}
}

func (d *GpuTypesDataSource) getClient() *client.RunPodClient {
	if d.client != nil {
		return d.client
	}
	apiKey := os.Getenv("RUNPOD_API_KEY")
	graphqlEndpoint := os.Getenv("RUNPOD_GRAPHQL_URL")
	if graphqlEndpoint == "" {
		graphqlEndpoint = "https://api.runpod.io/graphql"
	}
	restBaseURL := os.Getenv("RUNPOD_BASE_URL")
	if restBaseURL == "" {
		restBaseURL = "https://api.runpod.io/v2"
	}
	d.client = client.NewRunPodClient(apiKey, graphqlEndpoint, restBaseURL)
	return d.client
}

func (d *GpuTypesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "runpod_gpu_types"
}

func (d *GpuTypesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = GpuTypesDataSourceSchema(ctx)
}

func (d *GpuTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	client := d.getClient()

	// Use v2 REST endpoint: GET /v2/gpu
	url := client.RestBaseURL + "/gpu"

	reqHTTP, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.APIKey))

	httpClient := &http.Client{}
	respHTTP, err := httpClient.Do(reqHTTP)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to make API call: %v", err))
		return
	}
	defer respHTTP.Body.Close()

	// Handle non-200 responses
	if respHTTP.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(respHTTP.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to fetch GPU types (status %d): %s", respHTTP.StatusCode, string(body)))
		return
	}

	// Read and parse response body
	respBody, err := io.ReadAll(respHTTP.Body)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to read response: %v", err))
		return
	}

	var envelope map[string]interface{}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to parse response (status: %d): %s", respHTTP.StatusCode, string(respBody)))
		return
	}

	// Handle v2 response envelope: {data: {...}, meta: {...}, error: ...}
	var result map[string]interface{}
	if data, ok := envelope["data"].(map[string]interface{}); ok {
		result = data
	} else {
		result = envelope
	}

	// Parse GPU types list
	gpus, ok := result["gpus"].([]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Failed to parse GPU types from response - 'gpus' field missing or not an array")
		return
	}

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
}
