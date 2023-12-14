# Configure remote terraform backend for state.
terraform {
  backend "gcs" {
    bucket = "armored-witness-bucket-tfstate"
    prefix = "terraform/build_and_release/state"
  }
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
resource "google_storage_bucket" "armored_witness_firmware" {
  location                    = "EU"
  name                        = "armored-witness-firmware-1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "armored_witness_firmware_ci_1" {
  location                    = "EU"
  name                        = "armored-witness-firmware-ci-1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "armored_witness_firmware_log_1" {
  location                    = "US"
  name                        = "armored-witness-firmware-log-1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "armored_witness_firmware_log_ci_1" {
  location                    = "US"
  name                        = "armored-witness-firmware-log-ci-1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}

# KMS key rings
resource "google_kms_key_ring" "firmware_release_ci" {
  location = var.signing_keyring_location
  name     = "firmware-release-ci"
}
resource "google_kms_key_ring" "firmware_release_prod" {
  location = var.signing_keyring_location
  name     = "firmware-release-prod"
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

resource "google_kms_key_ring" "terraform_state" {
  name     = "armored-witness-bucket-tfstate"
  location = var.tf_state_location
}

resource "google_kms_crypto_key" "terraform_state_bucket" {
  name     = "terraform-state-bucket"
  key_ring = google_kms_key_ring.terraform_state.id
}

resource "google_storage_bucket" "terraform_state" {
  name          = "armored-witness-bucket-tfstate"
  force_destroy = false
  location      = var.tf_state_location
  storage_class = "STANDARD"
  versioning {
    enabled = true
  }
  encryption {
    default_kms_key_name = google_kms_crypto_key.terraform_state_bucket.id
  }
  uniform_bucket_level_access = true
}
# TODO(jayhou): add GCF stuff.