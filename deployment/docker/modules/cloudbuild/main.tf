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

resource "google_service_account" "cloudbuild_service_account" {
  account_id   = "cloudbuild-docker-${var.env}-sa"
  display_name = "Service Account for Docker CloudBuild (${var.env})"
}

resource "google_project_iam_member" "act_as" {
  project = var.project_id
  role    = "roles/iam.serviceAccountUser"
  member  = "serviceAccount:${google_service_account.cloudbuild_service_account.email}"
}

resource "google_project_iam_member" "service_agent" {
  project = var.project_id
  role    = "roles/cloudbuild.serviceAgent"
  member  = "serviceAccount:${google_service_account.cloudbuild_service_account.email}"
}

resource "google_project_iam_member" "logs_writer" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.cloudbuild_service_account.email}"
}

resource "google_project_iam_member" "artifact_registry_writer" {
  project = var.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${google_service_account.cloudbuild_service_account.email}"
}

