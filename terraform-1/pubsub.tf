# Pub/Sub Topic for audit logs
resource "google_pubsub_topic" "audit_logs_topic" {
  name    = var.pubsub_topic_name
  project = var.project_id

  labels = var.labels

  message_retention_duration = "86400s" # 24 hours

  depends_on = [google_project_service.required_apis]
}

# Pub/Sub Subscription with push to Cloud Run
resource "google_pubsub_subscription" "audit_logs_push_subscription" {
  name    = "${var.pubsub_topic_name}-push-sub"
  topic   = google_pubsub_topic.audit_logs_topic.name
  project = var.project_id

  labels = var.labels

  ack_deadline_seconds = 300
  message_retention_duration = "86400s"
  retain_acked_messages      = false

  # Push config will be updated after Cloud Run deployment
  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  depends_on = [google_project_service.required_apis]
}