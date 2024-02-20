terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "5.14.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# This will be configured by terragrunt when deploying
terraform {
  backend "gcs" {}
}

resource "google_artifact_registry_repository" "docker" {
  repository_id = "docker-${var.env}"
  location       = var.region
  description   = "docker images for armored witness tools"
  format        = "DOCKER"
}

locals {
  artifact_repo  = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.docker.name}"
  verifier_docker_address = "${local.artifact_repo}/verifybuild:latest"
}

resource "google_cloudbuild_trigger" "docker" {
  name            = "build-docker-${var.env}"
  location        = var.region
  # TODO(mhutchinson): this should be configured with a service account

  github {
    owner = "transparency-dev"
    name  = "armored-witness"
    push {
      branch = "^main$"
    }
  }

  # TODO(mhutchinson): Consider replacing this with https://github.com/ko-build/terraform-provider-ko
  build {
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
        "build",
        "-t", "${local.verifier_docker_address}",
        "-f", "./cmd/verify_build/Dockerfile",
        "."
      ]
    }
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
        "push",
        local.verifier_docker_address
      ]
    }
    options {
      logging = "CLOUD_LOGGING_ONLY"
    }
  }
}

