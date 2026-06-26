package resource_pod_action

import (
	"os"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/runpod/terraform-provider-runpod/internal/provider/client"
)

func NewPodActionResource() resource.Resource {
	return &PodActionResource{}
}

type PodActionResource struct{}

func (r *PodActionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "runpod_pod_action"
}

func (r *PodActionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PodActionResourceSchema(ctx)
}

func (r *PodActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config PodActionModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var query string
	action := config.Action.ValueString()

	switch action {
	case "stop":
		query = `
			mutation StopPod($podId: String!) {
				podStop(podId: $podId) {
					id
					status
				}
			}
		`
	case "resume":
		query = `
			mutation ResumePod($podId: String!) {
				podResume(podId: $podId) {
					id
					status
				}
			}
		`
	case "restart":
		query = `
			mutation RestartPod($podId: String!) {
				podRestart(podId: $podId) {
					id
					status
				}
			}
		`
	case "reset":
		query = `
			mutation ResetPod($podId: String!) {
				podReset(podId: $podId) {
					id
					status
				}
			}
		`
	case "terminate":
		query = `
			mutation TerminatePod($podId: String!) {
				podTerminate(podId: $podId) {
					id
					status
				}
			}
		`
	default:
		resp.Diagnostics.AddError("Invalid Action", "Action must be one of: stop, resume, restart, reset, terminate")
		return
	}

variables := map[string]interface{}{
			"podId": config.PodId.ValueString(),
		}

	apiKey := os.Getenv("RUNPOD_API_KEY")
	runpodClient := client.NewRunPodClient(apiKey, client.GetGraphQLEndpoint())
	result, err := runpodClient.Query(ctx, query, variables)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	var status string
	switch action {
	case "stop":
		if podStop, ok := result["podStop"].(map[string]interface{}); ok {
			if podStatus, ok := podStop["status"].(string); ok {
				status = podStatus
			}
		}
	case "resume":
		if podResume, ok := result["podResume"].(map[string]interface{}); ok {
			if podStatus, ok := podResume["status"].(string); ok {
				status = podStatus
			}
		}
	case "restart":
		if podRestart, ok := result["podRestart"].(map[string]interface{}); ok {
			if podStatus, ok := podRestart["status"].(string); ok {
				status = podStatus
			}
		}
	case "reset":
		if podReset, ok := result["podReset"].(map[string]interface{}); ok {
			if podStatus, ok := podReset["status"].(string); ok {
				status = podStatus
			}
		}
	case "terminate":
		if podTerminate, ok := result["podTerminate"].(map[string]interface{}); ok {
			if podStatus, ok := podTerminate["status"].(string); ok {
				status = podStatus
			}
		}
	}

	config.Status = types.StringValue(status)

	diags = resp.State.Set(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
}

func (r *PodActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *PodActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *PodActionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
