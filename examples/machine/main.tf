variable "runpod_api_key" {
  type        = string
  description = "RunPod API key"
  sensitive   = true
}

variable "gpu_type_id" {
  type        = string
  description = "GPU type ID (e.g., 'RTX 3090')"
  default     = "RTX 3090"
}

variable "gpu_count" {
  type        = number
  description = "Number of GPUs"
  default     = 1
}

variable "cpu_count" {
  type        = number
  description = "Number of CPUs"
  default     = 8
}

variable "memory_in_gb" {
  type        = number
  description = "Memory in GB"
  default     = 32
}

variable "disk_in_gb" {
  type        = number
  description = "Disk size in GB"
  default     = 50
}

variable "secure_cloud" {
  type        = bool
  description = "Use secure cloud"
  default     = false
}

variable "listed" {
  type        = bool
  description = "List machine for others"
  default     = true
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

resource "runpod_machine" "bid" {
  gpu_type_id      = var.gpu_type_id
  gpu_count        = var.gpu_count
  cpu_count        = var.cpu_count
  memory_in_gb     = var.memory_in_gb
  disk_in_gb       = var.disk_in_gb
  secure_cloud     = var.secure_cloud
  listed           = var.listed
  host_price_per_gpu = 0.50
}

output "machine_id" {
  description = "Machine ID"
  value       = runpod_machine.bid.id
}

output "machine_status" {
  description = "Machine status"
  value       = runpod_machine.bid.status
}
