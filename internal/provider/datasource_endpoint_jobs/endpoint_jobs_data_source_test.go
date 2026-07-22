package datasource_endpoint_jobs

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

func TestEndpointJobsDataSource_Schema(t *testing.T) {
	resource := NewEndpointJobsDataSource()
	ctx := context.Background()

	resp := &datasource.SchemaResponse{}
	resource.Schema(ctx, datasource.SchemaRequest{}, resp)

	assert.Nil(t, resp.Diagnostics)
	assert.NotNil(t, resp.Schema)
}
