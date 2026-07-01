package datasource_template

import (
	"os"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/runpod/terraform-provider-runpod/internal/provider/client"
)

func NewTemplateDataSource() datasource.DataSource {
	return &TemplateDataSource{}
}

type TemplateDataSource struct{}

func (d *TemplateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "runpod_template"
}

func (d *TemplateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = TemplateDataSourceSchema(ctx)
}

func (d *TemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config TemplateModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	query := `
		query GetTemplate($templateId: String!) {
			template(input: { templateId: $templateId }) {
				id
				name
				imageName
				category
				containerDiskInGb
				containerRegistryAuthId
				dockerEntrypoint
				dockerStartCmd
				env
				isPublic
				isServerless
				ports
				readme
				volumeInGb
				volumeMountPath
				earned
				isRunpod
				runtimeInMin
			}
		}
	`

	variables := map[string]interface{}{
		"templateId": config.Id.ValueString(),
	}

	apiKey := os.Getenv("RUNPOD_API_KEY")
	runpodClient := client.NewRunPodClient(apiKey, client.GetGraphQLEndpoint())
	result, err := runpodClient.Query(ctx, query, variables)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

 if template, ok := result["template"].(map[string]interface{}); ok {
    envMap := make(map[string]attr.Value)
    if val, ok := template["env"].(map[string]interface{}); ok {
      for key, v := range val {
        if strVal, ok := v.(string); ok {
          envMap[key] = types.StringValue(strVal)
        }
      }
    }

    dockerEntrypoint := make([]attr.Value, 0)
    if val, ok := template["dockerEntrypoint"].([]interface{}); ok {
      for _, v := range val {
        if vStr, ok := v.(string); ok {
          dockerEntrypoint = append(dockerEntrypoint, types.StringValue(vStr))
        }
      }
    }

    dockerStartCmd := make([]attr.Value, 0)
    if val, ok := template["dockerStartCmd"].([]interface{}); ok {
      for _, v := range val {
        if vStr, ok := v.(string); ok {
          dockerStartCmd = append(dockerStartCmd, types.StringValue(vStr))
        }
      }
    }

    ports := make([]attr.Value, 0)
    if val, ok := template["ports"].([]interface{}); ok {
      for _, v := range val {
        if vStr, ok := v.(string); ok {
          ports = append(ports, types.StringValue(vStr))
        }
      }
    }

    var name, imageName, category, containerRegistryAuthId, readme, volumeMountPath string
    var containerDiskInGb, volumeInGb, earned, runtimeInMin float64
    var isPublic, isServerless, isRunpod bool
    
    if v, ok := template["name"].(string); ok {
      name = v
    } else {
      resp.Diagnostics.AddError("API Error", "Field 'name' is missing or not a string in template response")
      return
    }
    
    if v, ok := template["imageName"].(string); ok {
      imageName = v
    } else {
      resp.Diagnostics.AddError("API Error", "Field 'imageName' is missing or not a string in template response")
      return
    }
    
    if v, ok := template["category"].(string); ok {
      category = v
    } else {
      resp.Diagnostics.AddError("API Error", "Field 'category' is missing or not a string in template response")
      return
    }
    
    if v, ok := template["containerDiskInGb"].(float64); ok {
      containerDiskInGb = v
    } else {
      resp.Diagnostics.AddError("API Error", "Field 'containerDiskInGb' is missing or not a float64 in template response")
      return
    }
    
    if v, ok := template["containerRegistryAuthId"].(string); ok {
      containerRegistryAuthId = v
    } else {
      resp.Diagnostics.AddError("API Error", "Field 'containerRegistryAuthId' is missing or not a string in template response")
      return
    }
    
    if v, ok := template["isPublic"].(bool); ok {
      isPublic = v
    } else {
      resp.Diagnostics.AddError("API Error", "Field 'isPublic' is missing or not a bool in template response")
      return
    }
    
    if v, ok := template["isServerless"].(bool); ok {
      isServerless = v
    } else {
      resp.Diagnostics.AddError("API Error", "Field 'isServerless' is missing or not a bool in template response")
      return
    }
    
    if v, ok := template["readme"].(string); ok {
      readme = v
    } else {
      resp.Diagnostics.AddError("API Error", "Field 'readme' is missing or not a string in template response")
      return
    }
    
    if v, ok := template["volumeInGb"].(float64); ok {
      volumeInGb = v
    } else {
      resp.Diagnostics.AddError("API Error", "Field 'volumeInGb' is missing or not a float64 in template response")
      return
    }
    
    if v, ok := template["volumeMountPath"].(string); ok {
      volumeMountPath = v
    } else {
      resp.Diagnostics.AddError("API Error", "Field 'volumeMountPath' is missing or not a string in template response")
      return
    }
    
    if v, ok := template["earned"].(float64); ok {
      earned = v
    } else {
      resp.Diagnostics.AddError("API Error", "Field 'earned' is missing or not a float64 in template response")
      return
    }
    
    if v, ok := template["isRunpod"].(bool); ok {
      isRunpod = v
    } else {
      resp.Diagnostics.AddError("API Error", "Field 'isRunpod' is missing or not a bool in template response")
      return
    }
    
    if v, ok := template["runtimeInMin"].(float64); ok {
      runtimeInMin = v
    } else {
      resp.Diagnostics.AddError("API Error", "Field 'runtimeInMin' is missing or not a float64 in template response")
      return
    }

    dockerEntrypointList, diags := types.ListValue(types.StringType, dockerEntrypoint)
    if diags.HasError() {
      resp.Diagnostics.Append(diags...)
      return
    }

    dockerStartCmdList, diags := types.ListValue(types.StringType, dockerStartCmd)
    if diags.HasError() {
      resp.Diagnostics.Append(diags...)
      return
    }

    envObj, diags := types.MapValue(types.StringType, envMap)
    if diags.HasError() {
      resp.Diagnostics.Append(diags...)
      return
    }

    portsList, diags := types.ListValue(types.StringType, ports)
    if diags.HasError() {
      resp.Diagnostics.Append(diags...)
      return
    }

    model := TemplateModel{
      Id:                      config.Id,
      Name:                    types.StringValue(name),
      ImageName:               types.StringValue(imageName),
      Category:                types.StringValue(category),
      ContainerDiskInGb:       types.Int64Value(int64(containerDiskInGb)),
      ContainerRegistryAuthId: types.StringValue(containerRegistryAuthId),
      DockerEntrypoint:        dockerEntrypointList,
      DockerStartCmd:          dockerStartCmdList,
      Env:                     envObj,
      IsPublic:                types.BoolValue(isPublic),
      IsServerless:            types.BoolValue(isServerless),
      Ports:                   portsList,
      Readme:                  types.StringValue(readme),
      VolumeInGb:              types.Int64Value(int64(volumeInGb)),
      VolumeMountPath:         types.StringValue(volumeMountPath),
      Earned:                  types.Float64Value(earned),
      IsRunpod:                types.BoolValue(isRunpod),
      RuntimeInMin:            types.Int64Value(int64(runtimeInMin)),
    }
    diags = resp.State.Set(ctx, &model)
    if diags.HasError() {
      resp.Diagnostics.Append(diags...)
      return
    }
  } else {
    resp.Diagnostics.AddError("API Error", "Template not found in response")
  }
}
