# Separate resource to configure push endpoint after Cloud Run is ready
resource "null_resource" "update_push_subscription" {
  triggers = {
    cloud_run_url = google_cloud_run_v2_service.audit_processor.uri
    subscription  = google_pubsub_subscription.audit_logs_push_subscription.name
  }

  provisioner "local-exec" {
    command = <<-EOT
      gcloud pubsub subscriptions modify-push-config ${google_pubsub_subscription.audit_logs_push_subscription.name} \
        --push-endpoint=${google_cloud_run_v2_service.audit_processor.uri} \
        --push-auth-service-account=${google_service_account.cloud_run_sa.email} \
        --project=${var.project_id}
    EOT
  }

  depends_on = [
    google_cloud_run_v2_service.audit_processor,
    google_pubsub_subscription.audit_logs_push_subscription
  ]
}