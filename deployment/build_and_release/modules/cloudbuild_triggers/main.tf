locals {
  cloudbuild_path = ({
    "prod" = "release/cloudbuild.yaml"
    "ci" = "release/cloudbuild_ci.yaml"
    "presubmit" = "release/cloudbuild_presubmit.yaml"
  })
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

    dynamic "push" {
      for_each = var.env == "prod" ? [1] : []
      content {
        tag = ".*" 
      }
    }
    dynamic "push" {
      for_each = var.env == "ci" ? [1] : []
      content {
        branch = "^main$"
      }
    }
    dynamic "pull_request" {
      for_each = var.env == "presubmit" ? [1] : []
      content {
        branch          = ".*"
        comment_control = "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY"
      }
    }
  }
 
  filename = local.cloudbuild_path[var.env]
}

