# Configure remote terraform backend for state.
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
resource "google_project_service" "cloudkms_googleapis_com" {
  service = "cloudkms.googleapis.com"
}
resource "google_project_service" "logging_googleapis_com" {
  service = "logging.googleapis.com"
}
resource "google_project_service" "serviceusage_googleapis_com" {
  service = "serviceusage.googleapis.com"
}
resource "google_project_service" "privateca_api" {
  service            = "privateca.googleapis.com"
  disable_on_destroy = false
}

# KMS key rings & data sources
resource "google_kms_key_ring" "hab" {
  location = var.signing_keyring_location
  name     = "hab-${var.env}"
}
data "google_kms_key_ring" "hab" {
  location = google_kms_key_ring.hab.location
  name     = google_kms_key_ring.hab.name
}

### KMS keys

# HAB CSF key & data sources for each of the SRK intermediates below.
#
# The resource creates the key within the keyring, and the data sections below
# ultimately provide a mechanism for getting the public key out from the HSM
# so that it can be certified by the leaf certificates of the HAB PKI below. 
resource "google_kms_crypto_key" "hab_csf" {
  for_each = toset(local.hab_intermediates)

  key_ring = google_kms_key_ring.hab.id
  name     = format("hab-csf%d-rev%d-%s", each.value, var.hab_revision, var.env)
  purpose  = "ASYMMETRIC_SIGN"
  version_template {
    algorithm        = format("RSA_SIGN_PKCS1_%d_SHA256", var.hab_keylength)
    protection_level = "HSM"
  }
}
data "google_kms_crypto_key" "hab_csf" {
  for_each = toset(local.hab_intermediates)

  name     = format("hab-csf%s-rev%d-%s", each.value, var.hab_revision, var.env)
  key_ring = data.google_kms_key_ring.hab.id

  depends_on = [
    google_kms_crypto_key.hab_csf
  ]
}
data "google_kms_crypto_key_version" "hab_csf" {
  for_each = toset(local.hab_intermediates)

  crypto_key = data.google_kms_crypto_key.hab_csf[each.key].id
}
# HAB IMG key & data sources for each of the SRK intermediates below.
resource "google_kms_crypto_key" "hab_img" {
  for_each = toset(local.hab_intermediates)

  key_ring = google_kms_key_ring.hab.id
  name     = format("hab-img%d-rev%d-%s", each.value, var.hab_revision, var.env)
  purpose  = "ASYMMETRIC_SIGN"
  version_template {
    algorithm        = format("RSA_SIGN_PKCS1_%d_SHA256", var.hab_keylength)
    protection_level = "HSM"
  }
}
data "google_kms_crypto_key" "hab_img" {
  for_each = toset(local.hab_intermediates)

  name     = format("hab-img%s-rev%d-%s", each.value, var.hab_revision, var.env)
  key_ring = data.google_kms_key_ring.hab.id

  depends_on = [
    google_kms_crypto_key.hab_img
  ]
}
data "google_kms_crypto_key_version" "hab_img" {
  for_each = toset(local.hab_intermediates)

  crypto_key = data.google_kms_crypto_key.hab_img[each.key].id
}

###########################################################################
## HAB Certificate Authority config.
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

# HAB CA pool
resource "google_privateca_ca_pool" "hab" {
  name     = format("aw-hab-ca-pool-rev0-%s", var.env)
  location = var.region
  tier     = "ENTERPRISE"
  publishing_options {
    publish_ca_cert = true
    publish_crl     = false
  }
  issuance_policy {
    baseline_values {
      ca_options {
      }
      key_usage {
        base_key_usage {
        }
        extended_key_usage {
        }
      }
    }
  }
}

# HAB Root CA authority
resource "google_privateca_certificate_authority" "hab_root" {
  pool                     = google_privateca_ca_pool.hab.name
  certificate_authority_id = format("hab-root-rev%d-%s", var.hab_revision, var.env)
  location                 = var.region
  lifetime                 = format("%ds", var.hab_pki_lifetime)
  deletion_protection      = true

  type = "SELF_SIGNED"
  config {
    subject_config {
      subject {
        organization        = "TrustFabric"
        organizational_unit = format("ArmoredWitness %s", upper(var.env))
        common_name         = format("ArmoredWitness Root %s", upper(var.env))
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
  key_spec {
    algorithm = format("RSA_PKCS1_%d_SHA256", var.hab_keylength)
  }
  lifecycle {
    ignore_changes = [
      lifetime,
    ]
    prevent_destroy = true
  }
}

locals {
  // This simply gives us a list we can use, in combination with the for_each meta attribute, to create
  // multiple instances of the subordinate CAs & certs below.
  hab_intermediates = [for i in range(1, 1 + var.hab_num_intermediates) : format("%s", i)]
}

# HAB SRK intermediates (one each for hab_intermediates above)
resource "google_privateca_certificate_authority" "hab_srk" {
  for_each = toset(local.hab_intermediates)

  pool                     = google_privateca_ca_pool.hab.name
  certificate_authority_id = format("hab-srk%s-rev%d-%s", each.value, var.hab_revision, var.env)
  location                 = var.region
  lifetime                 = format("%ds", var.hab_pki_lifetime)
  deletion_protection      = "true"

  type = "SUBORDINATE"
  subordinate_config {
    certificate_authority = google_privateca_certificate_authority.hab_root.name
  }
  config {
    subject_config {
      subject {
        organization        = "TrustFabric"
        organizational_unit = format("ArmoredWitness %s", upper(var.env))
        common_name         = format("ArmoredWitness SRK%s %s", each.value, upper(var.env))
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
  key_spec {
    algorithm = format("RSA_PKCS1_%d_SHA256", var.hab_keylength)
  }
  lifecycle {
    ignore_changes = [
      lifetime,
    ]
    prevent_destroy = true
  }
}

# HAB CSF cert for each of the SRK intermediates above.
resource "google_privateca_certificate" "hab_csf" {
  for_each = google_privateca_certificate_authority.hab_srk

  name                  = format("hab-csf%s-rev%d%s-%s", each.key, var.hab_revision, var.hab_leaf_minor, var.env)
  location              = var.region
  pool                  = each.value.pool
  certificate_authority = each.value.certificate_authority_id
  lifetime              = format("%ds", var.hab_pki_lifetime)
  config {
    subject_config {
      subject {
        organization        = "TrustFabric"
        organizational_unit = format("ArmoredWitness %s", upper(var.env))
        common_name         = format("ArmoredWitness SRK%s CSF %s", each.key, upper(var.env))
      }
    }
    x509_config {
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
      key    = base64encode(data.google_kms_crypto_key_version.hab_csf[each.key].public_key[0].pem)
    }
  }
  lifecycle {
    ignore_changes = [
      lifetime,
      config[0].x509_config[0].ca_options,
    ]
  }
}

# HAB IMG cert for each of the SRK intermediates above.
resource "google_privateca_certificate" "hab_img" {
  for_each = google_privateca_certificate_authority.hab_srk

  name                  = format("hab-img%s-rev%d%s-%s", each.key, var.hab_revision, var.hab_leaf_minor, var.env)
  location              = var.region
  pool                  = each.value.pool
  certificate_authority = each.value.certificate_authority_id
  lifetime              = format("%ds", var.hab_pki_lifetime)
  config {
    subject_config {
      subject {
        organization        = "TrustFabric"
        organizational_unit = format("ArmoredWitness %s", upper(var.env))
        common_name         = format("ArmoredWitness SRK%s IMG %s", each.key, upper(var.env))
      }
    }
    x509_config {
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
      key    = base64encode(data.google_kms_crypto_key_version.hab_img[each.key].public_key[0].pem)
    }
  }
  lifecycle {
    ignore_changes = [
      lifetime,
      config[0].x509_config[0].ca_options,
    ]
  }
}