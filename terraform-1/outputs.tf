output "project_id" {
  description = "GCP Project ID"
  value       = var.project_id
}

output "region" {
  description = "GCP Region"
  value       = var.region
}

output "cloud_run_url" {
  description = "URL of the Cloud Run service"
  value       = google_cloud_run_v2_service.audit_processor.uri
}

output "cloud_run_service_name" {
  description = "Name of the Cloud Run service"
  value       = google_cloud_run_v2_service.audit_processor.name
}

output "artifact_registry_url" {
  description = "Artifact Registry repository URL"
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.audit_processor_repo.repository_id}"
}

output "pubsub_topic_name" {
  description = "Pub/Sub topic name"
  value       = google_pubsub_topic.audit_logs_topic.name
}

output "pubsub_subscription_name" {
  description = "Pub/Sub subscription name"
  value       = google_pubsub_subscription.audit_logs_push_subscription.name
}

output "pubsub_subscription_id" {
  description = "Pub/Sub subscription ID"
  value       = google_pubsub_subscription.audit_logs_push_subscription.id
}



output "storage_bucket_name" {
  description = "Storage bucket name for audit logs"
  value       = google_storage_bucket.audit_logs_bucket.name
}

output "storage_bucket_url" {
  description = "Storage bucket URL"
  value       = google_storage_bucket.audit_logs_bucket.url
}

output "service_account_email" {
  description = "Service account email"
  value       = google_service_account.cloud_run_sa.email
}

output "log_sink_name" {
  description = "Log sink name"
  value       = google_logging_project_sink.audit_log_sink.name
}

output "secret_names" {
  description = "Secret Manager secret names"
  value = {
    tls_ca_cert     = google_secret_manager_secret.tls_ca_cert.secret_id
    tls_client_cert = google_secret_manager_secret.tls_client_cert.secret_id
    tls_client_key  = google_secret_manager_secret.tls_client_key.secret_id
  }
}

output "docker_build_command" {
  description = "Command to build and push Docker image"
  value       = "gcloud builds submit --tag ${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.audit_processor_repo.repository_id}/${var.service_name}:latest ."
}

output "deployment_commands" {
  description = "Commands for deployment"
  value = [
    "# 1. Authenticate Docker",
    "gcloud auth configure-docker ${var.region}-docker.pkg.dev",
    "# 2. Build and push image",
    "gcloud builds submit --tag ${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.audit_processor_repo.repository_id}/${var.service_name}:latest .",
    "# 3. Update Cloud Run service",
    "gcloud run services update ${var.service_name} --region=${var.region} --image=${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.audit_processor_repo.repository_id}/${var.service_name}:latest"
  ]
}