package resource_pod

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	client "github.com/runpod/terraform-provider-runpod/internal/provider/client"
)

func NewPodResource() resource.Resource {
	return &PodResource{}
}

type PodResource struct{}

func (r *PodResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "runpod_pod"
}

func (r *PodResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PodResourceSchema(ctx)
}

func (r *PodResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config PodModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	hasTemplateId := !config.TemplateId.IsNull() && config.TemplateId.ValueString() != ""
	hasImageName := !config.ImageName.IsNull() && config.ImageName.ValueString() != ""

	if hasTemplateId && hasImageName {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Cannot specify both template_id and image_name. Use template_id for templates, or image_name for direct image deployment.",
		)
		return
	}

	if !hasTemplateId && !hasImageName {
		resp.Diagnostics.AddError(
			"Missing Required Field",
			"Must specify either template_id or image_name.",
		)
		return
	}

	// Use REST API endpoint
	apiKey := os.Getenv("RUNPOD_API_KEY")
	if apiKey == "" {
		resp.Diagnostics.AddError("API Error", "RUNPOD_API_KEY environment variable must be set. Get your API key from https://runpod.io/console/user/settings")
		return
	}

	url := client.GetRestBaseURL() + "/pods"

	// Build the REST API request body
	body := map[string]interface{}{
		"gpuCount": int64(config.GpuCount.ValueInt64()),
		"name":     config.Name.ValueString(),
	}

	if hasTemplateId {
		body["templateId"] = config.TemplateId.ValueString()
	} else {
		body["imageName"] = config.ImageName.ValueString()
	}

	if !config.CloudType.IsNull() && config.CloudType.ValueString() != "" {
		body["cloudType"] = config.CloudType.ValueString()
	}

	if config.VolumeInGb.ValueFloat64() > 0 {
		body["volumeInGb"] = int64(config.VolumeInGb.ValueFloat64())
	}

	if !config.NetworkVolumeId.IsNull() && config.NetworkVolumeId.ValueString() != "" {
		body["networkVolumeId"] = config.NetworkVolumeId.ValueString()
	}

	if !config.DockerEntrypoint.IsNull() && len(config.DockerEntrypoint.Elements()) > 0 {
		entrypoint := make([]string, 0)
		for _, element := range config.DockerEntrypoint.Elements() {
			if strVal, ok := element.(types.String); ok {
				entrypoint = append(entrypoint, strVal.ValueString())
			}
		}
		if len(entrypoint) > 0 {
			body["dockerEntrypoint"] = entrypoint
		}
	}

	if !config.DockerStartCmd.IsNull() && len(config.DockerStartCmd.Elements()) > 0 {
		startCmd := make([]string, 0)
		for _, element := range config.DockerStartCmd.Elements() {
			if strVal, ok := element.(types.String); ok {
				startCmd = append(startCmd, strVal.ValueString())
			}
		}
		if len(startCmd) > 0 {
			body["dockerStartCmd"] = startCmd
		}
	}

	if !config.Interruptible.IsNull() {
		body["interruptible"] = config.Interruptible.ValueBool()
	}

	if !config.VolumeEncrypted.IsNull() {
		body["volumeEncrypted"] = config.VolumeEncrypted.ValueBool()
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to marshal request body: %v", err))
		return
	}

	reqHTTP, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	respHTTP, err := client.Do(reqHTTP)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to make API call: %v", err))
		return
	}
	defer respHTTP.Body.Close()

	respBody, err := io.ReadAll(respHTTP.Body)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to read response: %v", err))
		return
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to parse response (status: %d): %s", respHTTP.StatusCode, string(respBody)))
		return
	}

	// Extract the pod ID from response
	if podID, ok := result["id"].(string); ok {
		config.Id = types.StringValue(podID)
	} else {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to get pod ID from response: %v", result))
		return
	}

	if config.StartSsh.IsNull() {
		config.StartSsh = types.BoolValue(false)
	}
	if config.StartJupyter.IsNull() {
		config.StartJupyter = types.BoolValue(false)
	}

	// Set the state
	diags = resp.State.Set(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
}

func (r *PodResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PodModel
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	apiKey := os.Getenv("RUNPOD_API_KEY")
	if apiKey == "" {
		resp.Diagnostics.AddError("API Error", "RUNPOD_API_KEY environment variable must be set")
		return
	}

	url := client.GetRestBaseURL() + "/pods/" + state.Id.ValueString()

	reqHTTP, err := http.NewRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	respHTTP, err := client.Do(reqHTTP)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to make API call: %v", err))
		return
	}
	defer respHTTP.Body.Close()

	if respHTTP.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	respBody, err := io.ReadAll(respHTTP.Body)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to read response: %v", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to parse response (status: %d): %s", respHTTP.StatusCode, string(respBody)))
		return
	}

	if result == nil {
		resp.Diagnostics.AddError("API Error", "Empty response from API")
		return
	}

	if val, ok := result["desiredStatus"].(string); ok && val != "" {
		state.Status = types.StringValue(val)
	}
	if val, ok := result["createdAt"].(string); ok && val != "" {
		state.CreatedAt = types.StringValue(val)
	}
	if val, ok := result["machineId"].(string); ok && val != "" {
		state.MachineId = types.StringValue(val)
	}
	if val, ok := result["costPerHr"].(float64); ok {
		state.CostPerHr = types.Float64Value(val)
	}
	if val, ok := result["memoryInGb"].(float64); ok {
		state.MemoryInGb = types.Float64Value(val)
	}
	if val, ok := result["volumeInGb"].(float64); ok {
		state.VolumeInGb = types.Float64Value(val)
	}
	if val, ok := result["containerDiskInGb"].(float64); ok {
		state.ContainerDiskInGb = types.Int64Value(int64(val))
	}

	if result["templateId"] != nil {
		if val, ok := result["templateId"].(string); ok && val != "" {
			state.TemplateId = types.StringValue(val)
		}
	}

	if machine, ok := result["machine"].(map[string]interface{}); ok {
		if val, ok := machine["gpuTypeId"].(string); ok && val != "" {
			state.GpuTypeId = types.StringValue(val)
		}
		if v, ok := machine["secureCloud"].(bool); ok {
			if v {
				state.CloudType = types.StringValue("SECURE")
			} else {
				state.CloudType = types.StringValue("COMMUNITY")
			}
		}
	}

	if result["cloudType"] != nil {
		if val, ok := result["cloudType"].(string); ok && val != "" {
			state.CloudType = types.StringValue(val)
		}
	}

	if result["networkVolume"] != nil {
		if nv, ok := result["networkVolume"].(map[string]interface{}); ok {
			if id, ok := nv["id"].(string); ok && id != "" {
				state.NetworkVolumeId = types.StringValue(id)
			}
		}
	}

	if val, ok := result["dockerEntrypoint"].([]interface{}); ok {
		entrypointList := make([]attr.Value, 0)
		for _, v := range val {
			if vStr, ok := v.(string); ok {
				entrypointList = append(entrypointList, types.StringValue(vStr))
			}
		}
		if len(entrypointList) > 0 {
			state.DockerEntrypoint, diags = types.ListValue(types.StringType, entrypointList)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
		}
	}

	if val, ok := result["dockerStartCmd"].([]interface{}); ok {
		startCmdList := make([]attr.Value, 0)
		for _, v := range val {
			if vStr, ok := v.(string); ok {
				startCmdList = append(startCmdList, types.StringValue(vStr))
			}
		}
		if len(startCmdList) > 0 {
			state.DockerStartCmd, diags = types.ListValue(types.StringType, startCmdList)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
		}
	}

	if val, ok := result["interruptible"].(bool); ok {
		state.Interruptible = types.BoolValue(val)
	}

	if val, ok := result["volumeEncrypted"].(bool); ok {
		state.VolumeEncrypted = types.BoolValue(val)
	}

	diags = resp.State.Set(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
}

func (r *PodResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state PodModel
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var config PodModel
	diags = req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	apiKey := os.Getenv("RUNPOD_API_KEY")
	if apiKey == "" {
		resp.Diagnostics.AddError("API Error", "RUNPOD_API_KEY environment variable must be set")
		return
	}

	url := client.GetRestBaseURL() + "/pods/" + state.Id.ValueString()

	body := map[string]interface{}{}

	if !config.Name.IsNull() && config.Name.ValueString() != state.Name.ValueString() {
		body["name"] = config.Name.ValueString()
	}

	if !config.GpuCount.IsNull() && config.GpuCount.ValueInt64() != state.GpuCount.ValueInt64() {
		body["gpuCount"] = int64(config.GpuCount.ValueInt64())
	}

	if !config.CloudType.IsNull() && config.CloudType.ValueString() != state.CloudType.ValueString() {
		body["cloudType"] = config.CloudType.ValueString()
	}

	if !config.BidPerGpu.IsNull() {
		body["bidPerGpu"] = config.BidPerGpu.ValueFloat64()
	}

	if !config.DockerArgs.IsNull() && config.DockerArgs.ValueString() != state.DockerArgs.ValueString() {
		body["dockerArgs"] = config.DockerArgs.ValueString()
	}

	if !config.Env.IsNull() {
		envMap := make(map[string]interface{})
		for _, element := range config.Env.Elements() {
			if elementStr, ok := element.(types.String); ok {
				parts := strings.SplitN(elementStr.ValueString(), "=", 2)
				if len(parts) == 2 {
					envMap[parts[0]] = parts[1]
				}
			}
		}
		if len(envMap) > 0 {
			body["env"] = envMap
		}
	}

	if !config.Port.IsNull() && config.Port.ValueInt64() != state.Port.ValueInt64() {
		body["port"] = int64(config.Port.ValueInt64())
	}

	if !config.Ports.IsNull() && config.Ports.ValueString() != state.Ports.ValueString() {
		body["ports"] = config.Ports.ValueString()
	}

	if !config.StartSsh.IsNull() && config.StartSsh.ValueBool() != state.StartSsh.ValueBool() {
		body["startSsh"] = config.StartSsh.ValueBool()
	}

	if !config.StartJupyter.IsNull() && config.StartJupyter.ValueBool() != state.StartJupyter.ValueBool() {
		body["startJupyter"] = config.StartJupyter.ValueBool()
	}

	if !config.StopAfter.IsNull() && config.StopAfter.ValueString() != state.StopAfter.ValueString() {
		body["stopAfter"] = config.StopAfter.ValueString()
	}

	if !config.TerminateAfter.IsNull() && config.TerminateAfter.ValueString() != state.TerminateAfter.ValueString() {
		body["terminateAfter"] = config.TerminateAfter.ValueString()
	}

	if !config.VolumeInGb.IsNull() && config.VolumeInGb.ValueFloat64() != state.VolumeInGb.ValueFloat64() {
		body["volumeInGb"] = int64(config.VolumeInGb.ValueFloat64())
	}

	if !config.VolumeMountPath.IsNull() && config.VolumeMountPath.ValueString() != state.VolumeMountPath.ValueString() {
		body["volumeMountPath"] = config.VolumeMountPath.ValueString()
	}

	if len(body) == 0 {
		diags = resp.State.Set(ctx, &config)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
		return
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to marshal request body: %v", err))
		return
	}

	reqHTTP, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	httpClient := &http.Client{}
	respHTTP, err := httpClient.Do(reqHTTP)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to make API call: %v", err))
		return
	}
	defer respHTTP.Body.Close()

	respBody, err := io.ReadAll(respHTTP.Body)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to read response: %v", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to parse response (status: %d): %s", respHTTP.StatusCode, string(respBody)))
		return
	}

	if respHTTP.StatusCode != 200 && respHTTP.StatusCode != 201 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update pod (status: %d): %s", respHTTP.StatusCode, string(respBody)))
		return
	}

	config.Id = types.StringValue(result["id"].(string))

	diags = resp.State.Set(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
}

func (r *PodResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PodModel
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	apiKey := os.Getenv("RUNPOD_API_KEY")
	if apiKey == "" {
		resp.Diagnostics.AddError("API Error", "RUNPOD_API_KEY environment variable must be set")
		return
	}

	url := client.GetRestBaseURL() + "/pods/" + state.Id.ValueString()

	reqHTTP, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	respHTTP, err := client.Do(reqHTTP)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to make API call: %v", err))
		return
	}
	defer respHTTP.Body.Close()

	respBody, err := io.ReadAll(respHTTP.Body)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to read response: %v", err))
		return
	}

	if respHTTP.StatusCode != 204 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to delete pod (status: %d): %s", respHTTP.StatusCode, string(respBody)))
		return
	}
}
