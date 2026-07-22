package resource_endpoint_job

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func EndpointJobResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"endpoint_id": schema.StringAttribute{
				Required: true,
			},
			"input": schema.StringAttribute{
				Optional: true,
			},
			"template_id": schema.StringAttribute{
				Optional: true,
			},
			"worker_id": schema.StringAttribute{
				Optional: true,
			},
			"http_callback_url": schema.StringAttribute{
				Optional: true,
			},
			"http_callback_method": schema.StringAttribute{
				Optional: true,
			},
			"pause_logs": schema.BoolAttribute{
				Optional: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"duration_ms": schema.Int64Attribute{
				Computed: true,
			},
			"completed_at": schema.StringAttribute{
				Computed: true,
			},
			"output": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type EndpointJobModel struct {
	Id                 types.String `tfsdk:"id"`
	EndpointId         types.String `tfsdk:"endpoint_id"`
	Input              types.String `tfsdk:"input"`
	TemplateId         types.String `tfsdk:"template_id"`
	WorkerId           types.String `tfsdk:"worker_id"`
	HttpCallbackUrl    types.String `tfsdk:"http_callback_url"`
	HttpCallbackMethod types.String `tfsdk:"http_callback_method"`
	PauseLogs          types.Bool   `tfsdk:"pause_logs"`
	Status             types.String `tfsdk:"status"`
	CreatedAt          types.String `tfsdk:"created_at"`
	DurationMs         types.Int64  `tfsdk:"duration_ms"`
	CompletedAt        types.String `tfsdk:"completed_at"`
	Output             types.String `tfsdk:"output"`
}
