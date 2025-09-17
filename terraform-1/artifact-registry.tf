# Artifact Registry Repository
resource "google_artifact_registry_repository" "audit_processor_repo" {
  location      = var.region
  repository_id = "${var.service_name}-repo"
  description   = "Docker repository for ${var.service_name}"
  format        = "DOCKER"
  project       = var.project_id

  labels = var.labels

  cleanup_policies {
    id     = "keep-minimum-versions"
    action = "KEEP"
    
    most_recent_versions {
      keep_count = 10
    }
  }

  cleanup_policies {
    id     = "delete-old-versions"
    action = "DELETE"
    
    condition {
      older_than = "2592000s" # 30 days
    }
  }

  depends_on = [google_project_service.required_apis]
}
resource "google_artifact_registry_repository_iam_member" "cloud_run_access" {
  project    = var.project_id
  location   = google_artifact_registry_repository.audit_processor_repo.location
  repository = google_artifact_registry_repository.audit_processor_repo.name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${var.existing_service_account_email}"
}