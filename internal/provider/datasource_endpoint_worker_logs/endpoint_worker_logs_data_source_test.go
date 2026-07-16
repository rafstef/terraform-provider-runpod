package datasource_endpoint_worker_logs

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

func TestEndpointWorkerLogsDataSource_Schema(t *testing.T) {
	resource := NewEndpointWorkerLogsDataSource()
	ctx := context.Background()

	resp := &datasource.SchemaResponse{}
	resource.Schema(ctx, datasource.SchemaRequest{}, resp)

	assert.Nil(t, resp.Diagnostics)
	assert.NotNil(t, resp.Schema)
}
