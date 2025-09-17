# Cloud Run Service
resource "google_cloud_run_v2_service" "audit_processor" {
  name     = var.service_name
  location = var.region
  project  = var.project_id

  labels = var.labels

  template {
    service_account = google_service_account.cloud_run_sa.email

    scaling {
      max_instance_count = 100
      min_instance_count = 0
    }

    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.audit_processor_repo.repository_id}/${var.service_name}:latest"

      ports {
        container_port = 8080
      }

      # Environment variables
      env {
        name  = "PROJECT_ID"
        value = var.project_id
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
        name  = "PUBSUB_TOPIC"
        value = google_pubsub_topic.audit_logs_topic.name
      }



      resources {
        limits = {
          cpu    = "2"
          memory = "2Gi"
        }
        cpu_idle          = false
        startup_cpu_boost = true
      }

      # Health check
      startup_probe {
        http_get {
          path = "/health"
          port = 8080
        }
        initial_delay_seconds = 10
        timeout_seconds       = 5
        period_seconds        = 10
        failure_threshold     = 3
      }

      liveness_probe {
        http_get {
          path = "/health"
          port = 8080
        }
        initial_delay_seconds = 30
        timeout_seconds       = 5
        period_seconds        = 30
        failure_threshold     = 3
      }
    }

    timeout = "300s"

    # VPC connector if needed
    # vpc_access {
    #   connector = google_vpc_access_connector.connector.id
    #   egress    = "ALL_TRAFFIC"
    # }
  }

  traffic {
    percent = 100
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
  }

  depends_on = [
    google_artifact_registry_repository.audit_processor_repo
  ]
}

# IAM binding for Cloud Run invoker
resource "google_cloud_run_v2_service_iam_member" "invoker" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.audit_processor.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.cloud_run_sa.email}"
}

# Allow unauthenticated access for Pub/Sub push
resource "google_cloud_run_v2_service_iam_member" "public_invoker" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.audit_processor.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}