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
resource "google_project_service" "privateca_api" {
  service            = "privateca.googleapis.com"
  disable_on_destroy = false
}


# GCS buckets

# prod log rev 0
resource "google_storage_bucket" "armored_witness_firmware" {
  location                    = "EU"
  name                        = "armored-witness-firmware"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "armored_witness_firmware_log" {
  location                    = "US"
  name                        = "armored-witness-firmware-log"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}

# prod log rev 1
resource "google_storage_bucket" "armored_witness_firmware_1" {
  location                    = "EU"
  name                        = "armored-witness-firmware-1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "armored_witness_firmware_log_1" {
  location                    = "US"
  name                        = "armored-witness-firmware-log-1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}

# CI log rev 0
resource "google_storage_bucket" "armored_witness_firmware_ci" {
  location                    = "EU"
  name                        = "armored-witness-firmware-ci"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "armored_witness_firmware_log_ci" {
  location                    = "US"
  name                        = "armored-witness-firmware-log-ci"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}

# CI log rev 1
resource "google_storage_bucket" "armored_witness_firmware_ci_1" {
  location                    = "EU"
  name                        = "armored-witness-firmware-ci-1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "armored_witness_firmware_log_ci_1" {
  location                    = "US"
  name                        = "armored-witness-firmware-log-ci-1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}

# KMS key rings & data sources
resource "google_kms_key_ring" "firmware_release_ci" {
  location = var.signing_keyring_location
  name     = "firmware-release-ci"
}
resource "google_kms_key_ring" "firmware_release_prod" {
  location = var.signing_keyring_location
  name     = "firmware-release-prod"
}
resource "google_kms_key_ring" "hab_ci" {
  location = var.signing_keyring_location
  name     = "hab-ci"
}
data "google_kms_key_ring" "hab_ci" {
  location = google_kms_key_ring.hab_ci.location
  name     = google_kms_key_ring.hab_ci.name
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

# CI HAB CSF key & data sources for each of the SRK intermediates below.
#
# The resource creates the key within the keyring, and the data sections below
# ultimately provide a mechanism for getting the public key out from the HSM
# so that it can be certified by the leaf certificates of the HAB PKI below. 
resource "google_kms_crypto_key" "hab_csf_ci" {
  for_each = toset(local.hab_intermediates)

  key_ring = google_kms_key_ring.hab_ci.id
  name     = format("hab-csf%d-rev%d-ci", each.value, var.hab_ci_srk_revision)
  purpose  = "ASYMMETRIC_SIGN"
  version_template {
    algorithm        = format("RSA_SIGN_PKCS1_%d_SHA256", var.hab_keylength)
    protection_level = "HSM"
  }
}
data "google_kms_crypto_key" "hab_csf_ci" {
  for_each = toset(local.hab_intermediates)

  name     = format("hab-csf%s-rev%d-ci", each.value, var.hab_ci_srk_revision)
  key_ring = data.google_kms_key_ring.hab_ci.id

  depends_on = [
    google_kms_crypto_key.hab_csf_ci
  ]
}
data "google_kms_crypto_key_version" "hab_csf_ci" {
  for_each = toset(local.hab_intermediates)

  crypto_key = data.google_kms_crypto_key.hab_csf_ci[each.key].id
}
# CI HAB IMG key & data sources for each of the SRK intermediates below.
resource "google_kms_crypto_key" "hab_img_ci" {
  for_each = toset(local.hab_intermediates)

  key_ring = google_kms_key_ring.hab_ci.id
  name     = format("hab-img%d-rev%d-ci", each.value, var.hab_ci_srk_revision)
  purpose  = "ASYMMETRIC_SIGN"
  version_template {
    algorithm        = format("RSA_SIGN_PKCS1_%d_SHA256", var.hab_keylength)
    protection_level = "HSM"
  }
}
data "google_kms_crypto_key" "hab_img_ci" {
  for_each = toset(local.hab_intermediates)

  name     = format("hab-img%s-rev%d-ci", each.value, var.hab_ci_srk_revision)
  key_ring = data.google_kms_key_ring.hab_ci.id

  depends_on = [
    google_kms_crypto_key.hab_img_ci
  ]
}
data "google_kms_crypto_key_version" "hab_img_ci" {
  for_each = toset(local.hab_intermediates)

  crypto_key = data.google_kms_crypto_key.hab_img_ci[each.key].id
}

###########################################################################
## CI HAB Certificate Authority config.
## 
## This should construct a CA hierarchy as below for use with HAB signing:
##                                .--------.
##                                |  Root  |
##                                `--------'
##                                     |
##            ---------------------------------------------------
##            |               |                 |               |
##       .--------.       .--------.       .--------.       .--------.
##       |  SRK1  |       |  SRK2  |       |  SRK3  |       |  SRK4  |
##       `--------'       `--------'       `--------'       `--------'
##         /    \           /    \           /    \           /    \
##     .----.  .----.   .----.  .----.   .----.  .----.   .----.  .----.
##     |CSF1|  |IMG1|   |CSF2|  |IMG2|   |CSF3|  |IMG3|   |CSF4|  |IMG4|
##     `----'  `----'   `----'  `----'   `----'  `----'   `----'  `----'
###########################################################################

# CI HAB CA pool
resource "google_privateca_ca_pool" "hab_ci" {
  name     = "aw-hab-ca-pool-rev0-ci"
  location = "us-central1"
  tier     = "ENTERPRISE"
  publishing_options {
    publish_ca_cert = true
    publish_crl     = false
  }
  issuance_policy {
    baseline_values {
      ca_options {
        is_ca = true
      }
      key_usage {
        base_key_usage {
          cert_sign         = true
          crl_sign          = true
          digital_signature = true
        }
        extended_key_usage {
        }
      }
    }
  }
}

# CI HAB Root CA authority
resource "google_privateca_certificate_authority" "hab_root_ci" {
  pool                     = google_privateca_ca_pool.hab_ci.name
  certificate_authority_id = "hab-root-ci"
  location                 = "us-central1"
  config {
    subject_config {
      subject {
        organization        = "TrustFabric"
        organizational_unit = "ArmoredWitness CI"
        common_name         = "ArmoredWitness Root CI"
      }
    }
    x509_config {
      ca_options {
        # is_ca *MUST* be true for certificate authorities
        is_ca = true
      }
      key_usage {
        base_key_usage {
          # cert_sign and crl_sign *MUST* be true for certificate authorities
          cert_sign = true
          crl_sign  = true
        }
        extended_key_usage {
        }
      }
    }
  }
  type = "SELF_SIGNED"
  key_spec {
    algorithm = format("RSA_PKCS1_%d_SHA256", var.hab_keylength)
  }

  // REMOVE BEFORE SUBMIT:

  // Disable CA deletion related safe checks for easier cleanup.
  deletion_protection                    = false
  skip_grace_period                      = true
  ignore_active_certificates_on_deletion = true
}

locals {
  // This simply gives us a list we can use, in combination with the for_each meta attribute, to create
  // multiple instances of the subordinate CAs & certs below.
  hab_intermediates = [for i in range(1, 1 + var.hab_num_intermediates) : format("%s", i)]
}

# CI HAB SRK intermediates (one each for hab_intermediates above)
resource "google_privateca_certificate_authority" "hab_srk_ci" {
  for_each = toset(local.hab_intermediates)

  pool                     = google_privateca_ca_pool.hab_ci.name
  certificate_authority_id = format("hab-srk%s-rev%d-ci", each.value, var.hab_ci_srk_revision)
  location                 = "us-central1"

  //deletion_protection = "true"

  subordinate_config {
    certificate_authority = google_privateca_certificate_authority.hab_root_ci.name
  }
  config {
    subject_config {
      subject {
        organization        = "TrustFabric"
        organizational_unit = "ArmoredWitness CI"
        common_name         = format("ArmoredWitness SRK%s CI", each.value)
      }
    }
    x509_config {
      ca_options {
        is_ca = true
        # Force the sub CA to only issue leaf certs
        max_issuer_path_length = 0
      }
      key_usage {
        base_key_usage {
          cert_sign = true
          crl_sign  = true
        }
        extended_key_usage {
        }
      }
    }
  }
  lifetime = format("%ds", var.hab_pki_lifetime)
  key_spec {
    algorithm = format("RSA_PKCS1_%d_SHA256", var.hab_keylength)
  }
  type = "SUBORDINATE"

  // REMOVE BEFORE SUBMIT:

  // Disable CA deletion related safe checks for easier cleanup.
  deletion_protection                    = false
  skip_grace_period                      = true
  ignore_active_certificates_on_deletion = true
}

# CI HAB CSF cert for each of the SRK intermediates above.
resource "google_privateca_certificate" "hab_csf_ci" {
  for_each = google_privateca_certificate_authority.hab_srk_ci

  name                  = format("hab-csf%s-rev%d-ci", each.key, var.hab_ci_srk_revision)
  location              = "us-central1"
  pool                  = each.value.pool
  certificate_authority = each.value.certificate_authority_id
  lifetime              = format("%ds", var.hab_pki_lifetime)
  config {
    subject_config {
      subject {
        organization        = "TrustFabric"
        organizational_unit = "ArmoredWitness CI"
        common_name         = format("ArmoredWitness SRK%s CSF CI", each.key)
      }
    }
    x509_config {
      ca_options {
        is_ca = false
      }
      key_usage {
        base_key_usage {
          digital_signature = true
        }
        extended_key_usage {
        }
      }
    }
    public_key {
      format = "PEM"
      key    = base64encode(data.google_kms_crypto_key_version.hab_csf_ci[each.key].public_key[0].pem)
    }
  }
}

# CI HAB IMG cert for each of the SRK intermediates above.
resource "google_privateca_certificate" "hab_img_ci" {
  for_each = google_privateca_certificate_authority.hab_srk_ci

  name                  = format("hab-img%s-rev%d-ci", each.key, var.hab_ci_srk_revision)
  location              = "us-central1"
  pool                  = each.value.pool
  certificate_authority = each.value.certificate_authority_id
  lifetime              = format("%ds", var.hab_pki_lifetime)
  config {
    subject_config {
      subject {
        organization        = "TrustFabric"
        organizational_unit = "ArmoredWitness CI"
        common_name         = format("ArmoredWitness SRK%s IMG CI", each.key)
      }
    }
    x509_config {
      ca_options {
        is_ca = false
      }
      key_usage {
        base_key_usage {
          digital_signature = true
        }
        extended_key_usage {
        }
      }
    }
    public_key {
      format = "PEM"
      key    = base64encode(data.google_kms_crypto_key_version.hab_img_ci[each.key].public_key[0].pem)
    }
  }
}

############################################################
## Terraform state bucket
############################################################

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