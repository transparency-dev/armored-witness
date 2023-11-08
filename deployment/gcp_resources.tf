/**
 * Copyright 2022 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

## Project
provider "google-beta" {
  project = var.project_id
}

provider "google" {
  project = var.project_id
}

data "google_project" "project" {
  project_id = var.project_id
}

## Enable necessary APIs
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

### KMS key rings
resource "google_kms_key_ring" "firmware_release_ci" {
  location = "global"
  name     = "firmware-release-ci"
}
resource "google_kms_key_ring" "firmware_release_prod" {
  location = "global"
  name     = "firmware-release-prod"
}

# TODO(jayhou): This configuration cannot be applied right now because of the
# algorithm. Uncomment again when it is supported.
# ### KMS keys
# resource "google_kms_crypto_key" "bootloader_ci" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#   name                       = "bootloader-ci"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "recovery_ci" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#   name                       = "recovery-ci"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "trusted_applet_ci" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#   name                       = "trusted-applet-ci"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "trusted_os_1_ci" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#   name                       = "trusted-os-1-ci"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "trusted_os_2_ci" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#   name                       = "trusted-os-2-ci"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "ft_log_ci" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#   name                       = "ft-log-ci"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "bootloader_prod" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#   name                       = "bootloader-prod"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "recovery_prod" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#   name                       = "recovery-prod"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "trusted_applet_prod" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#   name                       = "trusted-applet-prod"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "trusted_os_prod" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#   name                       = "trusted-os-prod"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "ft_log_prod" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#   name                       = "ft-log-prod"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }

## GCS buckets
resource "google_storage_bucket" "armored_witness_firmware" {
  location                    = "EU"
  name                        = "armored-witness-firmware"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "armored_witness_firmware_ci" {
  location                    = "EU"
  name                        = "armored-witness-firmware-ci"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "armored_witness_firmware_log" {
  location                    = "US"
  name                        = "armored-witness-firmware-log"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "armored_witness_firmware_log_ci" {
  location                    = "US"
  name                        = "armored-witness-firmware-log-ci"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}

## TODO(jayhou): add GCF stuff.