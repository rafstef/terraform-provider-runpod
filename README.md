# RunPod Terraform Provider

A Terraform provider for managing RunPod infrastructure using the Terraform Plugin Framework.

## Setup

### Prerequisites

- Go 1.21 or higher
- Terraform 1.0 or higher
- RunPod API token

### Build the Provider

```bash
go build -o terraform-provider-runpod
```

### Configure Terraform

Create `~/.terraform.rc`:

```hcl
provider_installation {
  filesystem_paths {
    paths = ["."]
  }
}
```

Or use environment variable:

```bash
export TF_CLI_CONFIG_FILE=~/.terraform.rc
```

## Usage

### Basic Example

```hcl
terraform {
  required_providers {
    runpod = {
      source = "runpod/runpod"
    }
  }
}

# API key can be set via environment variable RUNPOD_API_KEY
# or in the provider configuration (not shown in this example)

resource "runpod_pod" "demo" {
  machine_id  = "your-machine-id"
  image_name  = "runpod/miniconda:py3.10-cuda11.8.0"
  gpu_count   = 1
  start_ssh   = true
}
```

### Environment Variable

Set your RunPod API key as an environment variable:

```bash
export RUNPOD_API_KEY="your-api-key-here"
```

Get your API key from [RunPod Console](https://runpod.io/console/user/settings)

### Examples Directory

- `examples/basic/` - Basic pod creation
- `examples/actions/` - Pod actions
- `examples/datasources/` - Data sources
- `examples/machine/` - Machine management
- `examples/monitoring/` - Pod monitoring

## Development

### Generate Provider Code

```bash
tfplugingen-framework generate all \
    --input terraform-provider-spec.json \
    --output internal/provider
```

### Directory Structure

```
terraform-provider/
├── internal/provider/          # Generated code
│   ├── provider_runpod/
│   ├── resource_pod/
│   ├── resource_pod_action/
│   ├── resource_machine/
│   ├── datasource_pod/
│   ├── datasource_machine/
│   └── ...
├── examples/                   # Example configurations
├── main.go                     # Provider entry point
├── plugin.go                   # Plugin interface
├── go.mod                      # Go dependencies
└── terraform-provider-spec.json # Provider specification
```

## API Documentation

- [RunPod API Docs](https://docs.runpod.io)
- [REST API Reference](https://rest.runpod.io/v1/docs)

## Provider Specification

The provider specification is defined in `terraform-provider-spec.json` and includes:

### Resources

- `runpod_pod` - Pod management
- `runpod_pod_action` - Pod actions
- `runpod_machine` - Machine management

### Data Sources

- `runpod_pod` - Pod information
- `runpod_machine` - Machine information
- `runpod_machines` - Machine listing
- `runpod_gpu_types` - GPU types
- `runpod_data_centers` - Data centers
- `runpod_user` - User info

## License

MIT
