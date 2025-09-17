variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-central1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "service_name" {
  description = "Cloud Run service name"
  type        = string
  default     = "audit-processor"
}

variable "pubsub_topic_name" {
  description = "Pub/Sub topic name"
  type        = string
  default     = "audit-logs"
}

variable "log_sink_name" {
  description = "Log sink name"
  type        = string
  default     = "audit-log-sink"
}

variable "existing_service_account_email" {
  description = "Email of existing service account to use"
  type        = string
}

variable "batch_size" {
  description = "Batch size for processing"
  type        = number
  default     = 100
}

variable "worker_count" {
  description = "Number of workers"
  type        = number
  default     = 10
}

variable "labels" {
  description = "Common labels for resources"
  type        = map(string)
  default = {
    project     = "audit-processor"
    environment = "dev"
    managed-by  = "terraform"
  }
}