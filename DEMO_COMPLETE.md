# RunPod Terraform Provider - Demo Setup Complete

## What Was Created

This repository now contains a complete demo setup for the RunPod Terraform provider:

### Generated Provider Code
- `internal/provider/` - Generated Go code for all resources and data sources
- 10 provider files (provider, 6 data sources, 3 resources)

### Example Configurations
- `examples/basic/` - Basic pod creation
- `examples/actions/` - Pod actions (stop/resume/terminate/reset)
- `examples/datasources/` - Querying GPU types, machines, data centers, users
- `examples/machine/` - Machine bidding and management
- `examples/monitoring/` - Pod monitoring

### Documentation
- `README.md` - Provider overview and usage guide
- `SETUP.md` - Step-by-step setup instructions
- `DEMO.md` - Demo examples and usage patterns
- `PROVIDER.md` - Provider resource and data source documentation
- `examples/README.md` - Example directory guide

## To Get Started

1. **Build the provider:**
   ```bash
   go build -o terraform-provider-runpod
   ```

2. **Configure Terraform:**
   ```bash
   echo 'provider_installation {
     filesystem_paths {
       paths = ["."]
     }
   }' > ~/.terraform.rc
   ```

3. **Update example configuration:**
   Edit `examples/basic/variables.tf` with your RunPod API key and machine ID

4. **Run the demo:**
   ```bash
   cd examples/basic
   terraform init
   terraform apply
   ```

## What You Can Do

- Create and manage RunPod pods
- Perform pod actions (stop, resume, terminate, reset)
- List available machines and GPU types
- Query data centers and user information
- Bid on machines for others to use

## Next Steps

1. Get your RunPod API key from https://runpod.io/console/user/settings
2. Follow `SETUP.md` for detailed instructions
3. Try the basic example to create your first pod
4. Explore other examples for advanced features
