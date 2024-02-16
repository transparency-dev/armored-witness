# Configure remote terraform backend for state.
# This will be configured by terragrunt when deploying.
terraform {
  backend "gcs" {}
}

# Project
provider "google" {
  project = var.project_id
}

data "google_project" "project" {
  project_id = var.project_id
}

# Enable necessary APIs
resource "google_project_service" "artifactregistry_googleapis_com" {
  service = "artifactregistry.googleapis.com"
}
resource "google_project_service" "cloudasset_googleapis_com" {
  service = "cloudasset.googleapis.com"
}
resource "google_project_service" "cloudbuild_googleapis_com" {
  service = "cloudbuild.googleapis.com"
}
resource "google_project_service" "cloudfunctions_googleapis_com" {
  service = "cloudfunctions.googleapis.com"
}
resource "google_project_service" "cloudkms_googleapis_com" {
  service = "cloudkms.googleapis.com"
}
resource "google_project_service" "compute_googleapis_com" {
  service = "compute.googleapis.com"
}
resource "google_project_service" "containerregistry_googleapis_com" {
  service = "containerregistry.googleapis.com"
}
resource "google_project_service" "logging_googleapis_com" {
  service = "logging.googleapis.com"
}
resource "google_project_service" "pubsub_googleapis_com" {
  service = "pubsub.googleapis.com"
}
resource "google_project_service" "serviceusage_googleapis_com" {
  service = "serviceusage.googleapis.com"
}
resource "google_project_service" "storage_api_googleapis_com" {
  service = "storage-api.googleapis.com"
}
resource "google_project_service" "storage_component_googleapis_com" {
  service = "storage-component.googleapis.com"
}
resource "google_project_service" "storage_googleapis_com" {
  service = "storage.googleapis.com"
}

# GCS buckets

# firmware rev 0
resource "google_storage_bucket" "firmware" {
  location                    = "EU"
  name                        = "armored-witness-firmware${var.bucket_env}"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "firmware_log" {
  location                    = "US"
  name                        = "armored-witness-firmware-log${var.bucket_env}"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}

# log rev 1
resource "google_storage_bucket" "firmware_1" {
  location                    = "EU"
  name                        = "armored-witness-firmware${var.bucket_env}-1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "firmware_log_1" {
  location                    = "US"
  name                        = "armored-witness-firmware-log${var.bucket_env}-1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}

# KMS key rings & data sources
resource "google_kms_key_ring" "firmware_release" {
  location = var.signing_keyring_location
  name     = "firmware-release-${var.env}"
}

# TODO(jayhou): This configuration cannot be applied right now because of the
# algorithm. Uncomment again when it is supported.
### KMS keys
#resource "google_kms_crypto_key" "bootloader_ci" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#  name     = "bootloader-ci"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "recovery_ci" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#  name     = "recovery-ci"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "trusted_applet_ci" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#  name     = "trusted-applet-ci"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "trusted_os_1_ci" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#  name     = "trusted-os-1-ci"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "trusted_os_2_ci" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#  name     = "trusted-os-2-ci"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "ft_log_ci" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#  name     = "ft-log-ci"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "bootloader_prod" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#  name     = "bootloader-prod"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "recovery_prod" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#  name     = "recovery-prod"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "trusted_applet_prod" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#  name     = "trusted-applet-prod"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "trusted_os_prod" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#  name     = "trusted-os-prod"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "ft_log_prod" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#  name     = "ft-log-prod"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}

# TODO(jayhou): add GCF stuff.
