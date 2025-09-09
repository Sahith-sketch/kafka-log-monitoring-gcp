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

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "audit-processor"
}

variable "pubsub_topic_name" {
  description = "Pub/Sub topic name"
  type        = string
  default     = "audit-logs"
}

variable "log_filter" {
  description = "Log sink filter"
  type        = string
  default     = "protoPayload.authenticationInfo.authoritySelector=\"kaiju-admin\" AND severity=\"DEFAULT\""
}

variable "cloud_run_image" {
  description = "Cloud Run container image"
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