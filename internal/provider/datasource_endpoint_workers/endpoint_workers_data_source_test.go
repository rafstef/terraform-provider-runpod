package datasource_endpoint_workers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

func TestEndpointWorkersDataSource_Schema(t *testing.T) {
	resource := NewEndpointWorkersDataSource()
	ctx := context.Background()

	resp := &datasource.SchemaResponse{}
	resource.Schema(ctx, datasource.SchemaRequest{}, resp)

	assert.Nil(t, resp.Diagnostics)
	assert.NotNil(t, resp.Schema)
}
