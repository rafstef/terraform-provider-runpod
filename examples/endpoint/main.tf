variable "runpod_api_key" {
  type        = string
  description = "RunPod API key"
  sensitive   = true
  default     = ""
}

variable "image_name" {
  type        = string
  description = "Container image name (v2 required)"
}

variable "gpu_type_id" {
  type        = string
  description = "GPU type ID (v2 required)"
}

variable "endpoint_name" {
  type        = string
  description = "Endpoint name"
  default     = "my-endpoint"
}

variable "workers_min" {
  type        = number
  description = "Minimum number of workers"
  default     = 0
}

variable "workers_max" {
  type        = number
  description = "Maximum number of workers"
  default     = 3
}

variable "idle_timeout" {
  type        = number
  description = "Idle timeout in minutes"
  default     = 5
}

variable "gpu_count" {
  type        = number
  description = "Number of GPUs per worker"
  default     = 1
}

terraform {
  required_providers {
    runpod = {
      source = "runpod/runpod"
    }
  }
}

provider "runpod" {
  api_key = var.runpod_api_key
}

resource "runpod_endpoint" "demo" {
  image_name   = var.image_name
  gpu_type_id  = var.gpu_type_id
  name         = var.endpoint_name
  workers_min  = var.workers_min
  workers_max  = var.workers_max
  idle_timeout = var.idle_timeout
  gpu_count    = var.gpu_count
}

output "endpoint_id" {
  description = "The endpoint ID"
  value       = runpod_endpoint.demo.id
}

output "endpoint_url" {
  description = "The endpoint URL for API calls"
  value       = "https://api.runpod.ai/v2/${runpod_endpoint.demo.id}/runsync"
}

output "endpoint_info" {
  description = "Endpoint configuration details"
  value = {
    name                 = runpod_endpoint.demo.name
    image_name           = runpod_endpoint.demo.image_name
    gpu_type_id          = runpod_endpoint.demo.gpu_type_id
    workers_min          = runpod_endpoint.demo.workers_min
    workers_max          = runpod_endpoint.demo.workers_max
    idle_timeout         = runpod_endpoint.demo.idle_timeout
    gpu_count            = runpod_endpoint.demo.gpu_count
    cloud_type           = runpod_endpoint.demo.cloud_type
    container_disk_in_gb = runpod_endpoint.demo.container_disk_in_gb
  }
}
