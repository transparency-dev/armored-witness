resource "google_compute_managed_ssl_certificate" "lb_default" {
  provider = google-beta
  name     = "transparency-dev-ssl-cert"
  project = var.project_id

  managed {
    domains = [var.serve_domain]
  }
}
