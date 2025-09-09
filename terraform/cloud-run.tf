resource "google_cloud_run_v2_service" "audit_processor" {
  name     = var.app_name
  location = var.region

  template {
    service_account = google_service_account.cloud_run_sa.email

    containers {
      image = var.cloud_run_image

      ports {
        container_port = 8080
      }

      env {
        name  = "BATCH_SIZE"
        value = tostring(var.batch_size)
      }

      env {
        name  = "WORKER_COUNT"
        value = tostring(var.worker_count)
      }

      env {
        name  = "STORAGE_BUCKET"
        value = google_storage_bucket.audit_logs.name
      }

      env {
        name  = "PROJECT_ID"
        value = var.project_id
      }

      resources {
        limits = {
          cpu    = "2"
          memory = "2Gi"
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }
  }

  depends_on = [google_project_service.required_apis]
}

resource "google_cloud_run_v2_service_iam_member" "pubsub_access" {
  project  = var.project_id
  location = google_cloud_run_v2_service.audit_processor.location
  name     = google_cloud_run_v2_service.audit_processor.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.pubsub_sa.email}"
}