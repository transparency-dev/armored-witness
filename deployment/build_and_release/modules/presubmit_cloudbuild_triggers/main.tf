terraform {
  backend "gcs" {}
}

resource "google_service_account" "builder" {
  account_id   = "cloudbuild-${var.env}"
  display_name = "Armored Witness ${var.env} Builder Service Account"
}

resource "google_cloudbuild_trigger" "applet_release" {
  location = "global"
  # service_account = google_service_account.builder.id

  github {
    owner = "transparency-dev"
    name  = "armored-witness-applet"

    pull_request {
      branch          = ".*"
      comment_control = "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY"
    }
  }
 
  filename = var.cloudbuild_path
}

