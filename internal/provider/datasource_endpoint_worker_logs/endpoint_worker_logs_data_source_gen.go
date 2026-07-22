package datasource_endpoint_worker_logs

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func EndpointWorkerLogsDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"endpoint_id": schema.StringAttribute{
				Required: true,
			},
			"worker_id": schema.StringAttribute{
				Required: true,
			},
			"logs": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type EndpointWorkerLogsDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	EndpointId types.String `tfsdk:"endpoint_id"`
	WorkerId   types.String `tfsdk:"worker_id"`
	Logs       types.String `tfsdk:"logs"`
}
