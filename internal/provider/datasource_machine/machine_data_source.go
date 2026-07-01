package datasource_machine

import (
	"os"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/runpod/terraform-provider-runpod/internal/provider/client"
)

func NewMachineDataSource() datasource.DataSource {
	return &MachineDataSource{}
}

type MachineDataSource struct{}

func (d *MachineDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "runpod_machine"
}

func (d *MachineDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = MachineDataSourceSchema(ctx)
}

func (d *MachineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MachineModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	query := `
		query GetMachine($machineId: String!) {
			machine(input: { machineId: $machineId }) {
				id
				name
				location
				listed
				gpuType {
					id
					name
				}
				gpuTotal
				gpuReserved
				cpuCount
				cpuTypeId
				memoryTotal
				memoryReserved
				diskTotal
				diskReserved
				secureCloud
				maintenanceMode
				verified
				hostPricePerGpu
				runpodIp
			}
		}
	`

variables := map[string]interface{}{
			"machineId": config.Id.ValueString(),
		}

	apiKey := os.Getenv("RUNPOD_API_KEY")
	runpodClient := client.NewRunPodClient(apiKey, client.GetGraphQLEndpoint())
	result, err := runpodClient.Query(ctx, query, variables)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	if machine, ok := result["machine"].(map[string]interface{}); ok {
		model := MachineModel{
			Id:              config.Id,
			Name:            types.StringValue(machine["name"].(string)),
			Location:        types.StringValue(machine["location"].(string)),
			Listed:          types.BoolValue(machine["listed"].(bool)),
			GpuTypeId:       types.StringValue(machine["gpuType"].(map[string]interface{})["id"].(string)),
			GpuTotal:        types.Int64Value(int64(machine["gpuTotal"].(float64))),
			GpuReserved:     types.Int64Value(int64(machine["gpuReserved"].(float64))),
			CpuCount:        types.Int64Value(int64(machine["cpuCount"].(float64))),
			CpuTypeId:       types.StringValue(machine["cpuTypeId"].(string)),
			MemoryTotal:     types.Int64Value(int64(machine["memoryTotal"].(float64))),
			MemoryReserved:  types.Int64Value(int64(machine["memoryReserved"].(float64))),
			DiskTotal:       types.Int64Value(int64(machine["diskTotal"].(float64))),
			DiskReserved:    types.Int64Value(int64(machine["diskReserved"].(float64))),
			SecureCloud:     types.BoolValue(machine["secureCloud"].(bool)),
			MaintenanceMode: types.BoolValue(machine["maintenanceMode"].(bool)),
			Verified:        types.BoolValue(machine["verified"].(bool)),
			HostPricePerGpu: types.Float64Value(machine["hostPricePerGpu"].(float64)),
			RunpodIp:        types.StringValue(machine["runpodIp"].(string)),
		}
		diags = resp.State.Set(ctx, &model)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	} else {
		resp.Diagnostics.AddError("API Error", "Machine not found in response")
	}
}
