terraform {
  backend "gcs" {}
}

resource "google_service_account" "builder" {
  account_id   = "cloudbuild-${var.env}"
  display_name = "Armored Witness ${var.env} Builder Service Account"
}

resource "google_cloudbuild_trigger" "release" {
  for_each = var.build_components

  location = "global"
  # TODO(jayhou): uncomment this once the service account is created and permissions are granted.
  # service_account = google_service_account.builder.id

  github {
    owner = "transparency-dev"
    name  = "${each.value.repo}"

    pull_request {
      branch          = ".*"
      comment_control = "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY"
    }
  }
 
  filename = "${each.value.cloudbuild_path}"
}

