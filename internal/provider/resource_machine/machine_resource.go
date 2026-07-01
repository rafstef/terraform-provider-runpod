package resource_machine

import (
	"os"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/runpod/terraform-provider-runpod/internal/provider/client"
)

func NewMachineResource() resource.Resource {
	return &MachineResource{}
}

type MachineResource struct {
	client *client.RunPodClient
}

func (r *MachineResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*client.RunPodClient)
	}
}

func (r *MachineResource) getClient() *client.RunPodClient {
	if r.client != nil {
		return r.client
	}
	apiKey := os.Getenv("RUNPOD_API_KEY")
	graphqlEndpoint := os.Getenv("RUNPOD_GRAPHQL_URL")
	if graphqlEndpoint == "" {
		graphqlEndpoint = "https://api.runpod.io/graphql"
	}
	restBaseURL := os.Getenv("RUNPOD_BASE_URL")
	if restBaseURL == "" {
		restBaseURL = "https://rest.runpod.io/v1"
	}
	r.client = client.NewRunPodClient(apiKey, graphqlEndpoint, restBaseURL)
	return r.client
}

func (r *MachineResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "runpod_machine"
}

func (r *MachineResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = MachineResourceSchema(ctx)
}

func (r *MachineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config MachineModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	query := `
		mutation CreateMachine($input: CreateMachineInput!) {
			machineAdd(input: $input) {
				id
				name
				description
				gpuCount
				gpuType
				cpuCount
				memoryInGb
				diskSizeInGb
				region
				priceHourly
				priceMonthly
				status
				createdAt
				updatedAt
			}
		}
	`

variables := map[string]interface{}{
			"input": map[string]interface{}{
				"name":         config.Name.ValueString(),
				"description":  config.Name.ValueString(),
				"gpuCount":     config.GpuCount.ValueInt64(),
				"gpuType":      config.GpuTypeId.ValueString(),
				"cpuCount":     config.CpuCount.ValueInt64(),
				"memoryInGb":   config.MemoryInGb.ValueInt64(),
				"diskSizeInGb": config.DiskInGb.ValueInt64(),
				"region":       config.Location.ValueString(),
				"secureCloud":  config.SecureCloud.ValueBool(),
				"listed":       config.Listed.ValueBool(),
			},
		}

	result, err := r.getClient().Query(ctx, query, variables)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	if machineAdd, ok := result["machineAdd"].(map[string]interface{}); ok {
		if machineID, ok := machineAdd["id"].(string); ok {
			config.Id = types.StringValue(machineID)
		} else {
			resp.Diagnostics.AddError("API Error", "Failed to get machine ID from response")
			return
		}
	} else {
		resp.Diagnostics.AddError("API Error", "machineAdd not in response")
		return
	}

	diags = resp.State.Set(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
}

func (r *MachineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var config MachineModel
	diags := req.State.Get(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	query := `
		query GetMachine($machineId: String!) {
			machine(input: { machineId: $machineId }) {
				id
				name
				description
				gpuCount
				gpuType
				cpuCount
				memoryInGb
				diskSizeInGb
				region
				priceHourly
				priceMonthly
				status
				createdAt
				updatedAt
				listed
				location
				secureCloud
				maintenanceMode
				verified
				hostPricePerGpu
				diskTotal
				diskReserved
				memoryTotal
				memoryReserved
				gpuTotal
				gpuReserved
				cpuTypeId
				runpodIp
			}
		}
	`

variables := map[string]interface{}{
			"machineId": config.Id.ValueString(),
		}

	result, err := r.getClient().Query(ctx, query, variables)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	if machine, ok := result["machine"].(map[string]interface{}); ok {
		var name, gpuType string
		
		if v, ok := machine["name"].(string); ok {
			name = v
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'name' is missing or not a string in machine response")
			return
		}
		
		if v, ok := machine["gpuCount"].(float64); ok {
			config.GpuCount = types.Int64Value(int64(v))
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'gpuCount' is missing or not a float64 in machine response")
			return
		}
		
		if v, ok := machine["gpuType"].(string); ok {
			gpuType = v
			config.GpuTypeId = types.StringValue(v)
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'gpuType' is missing or not a string in machine response")
			return
		}
		
		if v, ok := machine["cpuCount"].(float64); ok {
			config.CpuCount = types.Int64Value(int64(v))
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'cpuCount' is missing or not a float64 in machine response")
			return
		}
		
		if v, ok := machine["memoryInGb"].(float64); ok {
			config.MemoryInGb = types.Int64Value(int64(v))
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'memoryInGb' is missing or not a float64 in machine response")
			return
		}
		
		if v, ok := machine["diskSizeInGb"].(float64); ok {
			config.DiskInGb = types.Int64Value(int64(v))
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'diskSizeInGb' is missing or not a float64 in machine response")
			return
		}
		
		if v, ok := machine["region"].(string); ok {
			config.Location = types.StringValue(v)
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'region' is missing or not a string in machine response")
			return
		}
		
		if v, ok := machine["listed"].(bool); ok {
			config.Listed = types.BoolValue(v)
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'listed' is missing or not a bool in machine response")
			return
		}
		
		if v, ok := machine["secureCloud"].(bool); ok {
			config.SecureCloud = types.BoolValue(v)
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'secureCloud' is missing or not a bool in machine response")
			return
		}
		
		if v, ok := machine["maintenanceMode"].(bool); ok {
			config.MaintenanceMode = types.BoolValue(v)
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'maintenanceMode' is missing or not a bool in machine response")
			return
		}
		
		if v, ok := machine["verified"].(bool); ok {
			config.Verified = types.BoolValue(v)
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'verified' is missing or not a bool in machine response")
			return
		}
		
		if v, ok := machine["hostPricePerGpu"].(float64); ok {
			config.HostPricePerGpu = types.Float64Value(v)
		} else {
			resp.Diagnostics.AddError("API Error", "Field 'hostPricePerGpu' is missing or not a float64 in machine response")
			return
		}
		
		config.Name = types.StringValue(name)
		config.GpuTypeId = types.StringValue(gpuType)
	} else {
		resp.Diagnostics.AddError("API Error", "Machine not found in response")
		return
	}

	diags = resp.State.Set(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
}

func (r *MachineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var config MachineModel
	diags := req.Plan.Get(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	query := `
		mutation EditMachine($input: EditMachineInput!) {
			machineEditName(input: $input) {
				id
				name
				description
				gpuCount
				gpuType
				cpuCount
				memoryInGb
				diskSizeInGb
				region
				listed
			}
		}
	`

variables := map[string]interface{}{
			"input": map[string]interface{}{
				"id":           config.Id.ValueString(),
				"name":         config.Name.ValueString(),
				"description":  config.Name.ValueString(),
				"gpuCount":     config.GpuCount.ValueInt64(),
				"gpuType":      config.GpuTypeId.ValueString(),
				"cpuCount":     config.CpuCount.ValueInt64(),
				"memoryInGb":   config.MemoryInGb.ValueInt64(),
				"diskSizeInGb": config.DiskInGb.ValueInt64(),
				"region":       config.Location.ValueString(),
				"listed":       config.Listed.ValueBool(),
			},
		}

	_, err := r.getClient().Query(ctx, query, variables)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
}

func (r *MachineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var config MachineModel
	diags := req.State.Get(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	query := `
		mutation DeleteMachine($machineId: String!) {
			machineDelete(machineId: $machineId) {
				id
				status
			}
		}
	`

variables := map[string]interface{}{
			"machineId": config.Id.ValueString(),
		}

	_, err := r.getClient().Query(ctx, query, variables)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}
}
