# RunPod PyTorch Pod with RTX 4090

terraform {
  required_providers {
    runpod = {
      source = "registry.terraform.io/runpod/runpod"
    }
  }
}

# API key must be set via RUNPOD_API_KEY environment variable
# Run: export RUNPOD_API_KEY="your-api-key-here"
# before running: terraform init && terraform apply

resource "runpod_pod" "pytorch_experiment" {
  # Use template_id to deploy using a specific template
  template_id   = var.template_id
  image_name    = var.image_name
  gpu_count     = 1
  name          = "pytorch-experiment"
  start_ssh     = true
  start_jupyter = true
  volume_in_gb  = 10
}

# Output pod details
output "pod_id" {
  value = runpod_pod.pytorch_experiment.id
  description = "The pod ID for your PyTorch experiment"
}

output "pod_status" {
  value = runpod_pod.pytorch_experiment.status
  description = "Current pod status"
}
