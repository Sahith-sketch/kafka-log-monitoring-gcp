resource "google_pubsub_topic" "audit_logs" {
  name = var.pubsub_topic_name

  depends_on = [google_project_service.required_apis]
}

resource "google_pubsub_subscription" "audit_logs_push" {
  name  = "${var.pubsub_topic_name}-push-sub"
  topic = google_pubsub_topic.audit_logs.name

  push_config {
    push_endpoint = google_cloud_run_v2_service.audit_processor.uri

    oidc_token {
      service_account_email = google_service_account.pubsub_sa.email
    }
  }

  ack_deadline_seconds = 600

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  depends_on = [google_cloud_run_v2_service.audit_processor]
}