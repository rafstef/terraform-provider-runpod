package datasource_endpoint_job_logs

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

func TestEndpointJobLogsDataSource_Schema(t *testing.T) {
	resource := NewEndpointJobLogsDataSource()
	ctx := context.Background()

	resp := &datasource.SchemaResponse{}
	resource.Schema(ctx, datasource.SchemaRequest{}, resp)

	assert.Nil(t, resp.Diagnostics)
	assert.NotNil(t, resp.Schema)
}
