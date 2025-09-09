resource "google_logging_project_sink" "audit_logs_sink" {
  name        = "${var.app_name}-audit-sink"
  destination = "pubsub.googleapis.com/projects/${var.project_id}/topics/${google_pubsub_topic.audit_logs.name}"
  filter      = var.log_filter

  unique_writer_identity = true

  depends_on = [google_pubsub_topic.audit_logs]
}

resource "google_pubsub_topic_iam_member" "log_sink_publisher" {
  topic  = google_pubsub_topic.audit_logs.name
  role   = "roles/pubsub.publisher"
  member = google_logging_project_sink.audit_logs_sink.writer_identity
}