package datasource_endpoint_job_logs

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func EndpointJobLogsDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"endpoint_id": schema.StringAttribute{
				Required: true,
			},
			"job_id": schema.StringAttribute{
				Required: true,
			},
			"logs": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type EndpointJobLogsDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	EndpointId types.String `tfsdk:"endpoint_id"`
	JobId      types.String `tfsdk:"job_id"`
	Logs       types.String `tfsdk:"logs"`
}
