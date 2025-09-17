# Secret for TLS CA Certificate
resource "google_secret_manager_secret" "tls_ca_cert" {
  secret_id = "${var.service_name}-tls-ca-cert"
  project   = var.project_id

  labels = var.labels

  replication {
    auto {}
  }

  depends_on = [google_project_service.required_apis]
}

resource "google_secret_manager_secret_version" "tls_ca_cert_version" {
  secret      = google_secret_manager_secret.tls_ca_cert.id
  secret_data = file("${path.module}/../certs/ca.crt")
}

# Secret for TLS Client Certificate
resource "google_secret_manager_secret" "tls_client_cert" {
  secret_id = "${var.service_name}-tls-client-cert"
  project   = var.project_id

  labels = var.labels

  replication {
    auto {}
  }

  depends_on = [google_project_service.required_apis]
}

resource "google_secret_manager_secret_version" "tls_client_cert_version" {
  secret      = google_secret_manager_secret.tls_client_cert.id
  secret_data = file("${path.module}/../certs/cert.pem")
}

# Secret for TLS Client Key
resource "google_secret_manager_secret" "tls_client_key" {
  secret_id = "${var.service_name}-tls-client-key"
  project   = var.project_id

  labels = var.labels

  replication {
    auto {}
  }

  depends_on = [google_project_service.required_apis]
}

resource "google_secret_manager_secret_version" "tls_client_key_version" {
  secret      = google_secret_manager_secret.tls_client_key.id
  secret_data = file("${path.module}/../certs/cert.key")
}