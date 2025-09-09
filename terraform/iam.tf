resource "google_service_account" "cloud_run_sa" {
  account_id   = "${var.app_name}-sa"
  display_name = "Service Account for ${var.app_name}"
  description  = "Service account for Cloud Run service"
}

resource "google_project_iam_member" "cloud_run_permissions" {
  for_each = toset([
    "roles/pubsub.subscriber",
    "roles/storage.objectAdmin",
    "roles/logging.logWriter"
  ])

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.cloud_run_sa.email}"
}

resource "google_service_account" "pubsub_sa" {
  account_id   = "${var.app_name}-pubsub-sa"
  display_name = "Pub/Sub Service Account for ${var.app_name}"
  description  = "Service account for Pub/Sub to invoke Cloud Run"
}

resource "google_project_iam_member" "pubsub_invoker" {
  project = var.project_id
  role    = "roles/run.invoker"
  member  = "serviceAccount:${google_service_account.pubsub_sa.email}"
}