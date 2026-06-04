variable "runpod_api_key" {
  type        = string
  description = "RunPod API key"
  sensitive   = true
}

variable "image_name" {
  type        = string
  description = "Docker image name for the pod"
  default     = "runpod/pytorch:2.1.1-py3.10-cuda12.1.1-devel-ubuntu22.04"
}

variable "template_id" {
  type        = string
  description = "Template ID for the pod (get from https://www.runpod.io/console/templates)"
  default     = ""
}

variable "gpu_type_id" {
  type        = string
  description = "GPU type ID (e.g., 'NVIDIA GeForce RTX 4090') - for auto-selection"
  default     = ""
}
