# RunPod Terraform Provider Setup

This document provides a step-by-step guide to get the RunPod Terraform provider demo up and running.

## Quick Start

### 1. Build the Provider

```bash
cd /Users/books/repos/terraform-provider
go build -o terraform-provider-runpod
```

### 2. Configure Terraform

```bash
echo 'provider_installation {
  filesystem_paths {
    paths = ["."]
  }
}' > ~/.terraform.rc
```

### 3. Update Example Configuration

Edit `examples/basic/variables.tf` and replace:
- `YOUR_API_KEY_HERE` with your RunPod API key
- `YOUR_MACHINE_ID_HERE` with a machine ID from your RunPod account

### 4. Run the Demo

```bash
cd examples/basic
terraform init
terraform apply
```

## Detailed Setup

### Get RunPod API Key

1. Visit https://runpod.io/console/user/settings
2. Copy your API key
3. Store it securely (use environment variable or update variables.tf)

### Find Available Machines

Use the datasources demo to find available machines:

```bash
cd examples/datasources
terraform apply -var="runpod_api_key=your-key"
```

This will list all available machines you can deploy pods to.

### Create a Pod

Once you have a machine ID:

```bash
cd examples/basic
terraform apply \
  -var="runpod_api_key=your-key" \
  -var="machine_id=your-machine-id"
```

### Monitor Your Pod

```bash
cd examples/monitoring
terraform apply \
  -var="runpod_api_key=your-key" \
  -var="pod_id=the-pod-id-from-previous-step"
```

### Clean Up

```bash
cd examples/actions
terraform apply \
  -var="runpod_api_key=your-key" \
  -var="pod_id=your-pod-id" \
  -var="action=terminate"
```

## Troubleshooting

### Provider Not Found

Ensure the provider binary is in your current directory and `~/.terraform.rc` is configured correctly.

### No Machines Available

You need to list machines in your RunPod console before they appear in the API.

### Authentication Errors

Verify your API key is correct and has the necessary permissions.

## Next Steps

- Explore other examples in the `examples/` directory
- Modify the provider specification in `terraform-provider-spec.json`
- Implement custom logic in the generated provider code
- Contribute to the provider on GitHub
