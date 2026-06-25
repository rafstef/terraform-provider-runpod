variable "runpod_api_key" {
  type        = string
  description = "RunPod API key"
  sensitive   = true
  default     = ""
}

variable "machine_id" {
  type        = string
  description = "Machine ID to deploy pod on"
  default     = ""
}

variable "image_name" {
  type        = string
  description = "Docker image name"
  default     = "runpod/pytorch:1.0.7-cu1281-torch291-ubuntu2404"
}

variable "pod_name" {
  type        = string
  description = "Pod name"
  default     = "demo-pod-with-volume"
}

variable "gpu_count" {
  type        = number
  description = "Number of GPUs"
  default     = 1
}

variable "volume_size_gb" {
  type        = number
  description = "Volume size in GB"
  default     = 10
}

variable "volume_mount_path" {
  type        = string
  description = "Volume mount path"
  default     = "/workspace/data"
}

variable "start_ssh" {
  type        = bool
  description = "Start SSH on boot"
  default     = true
}

variable "start_jupyter" {
  type        = bool
  description = "Start Jupyter notebook on boot"
  default     = true
}
