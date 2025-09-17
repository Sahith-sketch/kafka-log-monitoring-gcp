# Log sink to capture audit logs and send to Pub/Sub
resource "google_logging_project_sink" "audit_log_sink" {
  name        = var.log_sink_name
  destination = "pubsub.googleapis.com/projects/${var.project_id}/topics/${google_pubsub_topic.audit_logs_topic.name}"
  project     = var.project_id

  # Filter for kaiju-admin authority selector and DEFAULT severity
  filter = "protoPayload.authenticationInfo.authoritySelector=\"kaiju-admin\" AND severity=\"DEFAULT\""

  unique_writer_identity = true

  depends_on = [google_pubsub_topic.audit_logs_topic]
}

# Grant the log sink service account permission to publish to Pub/Sub
resource "google_project_iam_member" "log_sink_publisher" {
  project = var.project_id
  role    = "roles/pubsub.publisher"
  member  = google_logging_project_sink.audit_log_sink.writer_identity
}

# Additional log sink for all audit logs (backup/monitoring)
resource "google_logging_project_sink" "all_audit_logs_sink" {
  name        = "${var.log_sink_name}-all"
  destination = "storage.googleapis.com/${google_storage_bucket.audit_logs.name}"
  project     = var.project_id

  # Capture all audit logs
  filter = "protoPayload.@type=\"type.googleapis.com/google.cloud.audit.AuditLog\""

  unique_writer_identity = true

  depends_on = [google_storage_bucket.audit_logs]
}

# Grant the backup log sink service account permission to write to storage
resource "google_project_iam_member" "backup_log_sink_writer" {
  project = var.project_id
  role    = "roles/storage.objectCreator"
  member  = google_logging_project_sink.all_audit_logs_sink.writer_identity
}