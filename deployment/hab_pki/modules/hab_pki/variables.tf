variable "project_id" {
  description = "The project ID to host the cluster in"
}

variable "region" {
  description = "The GCP region to host the resources in"
}

variable "env" {
  description = "Unique id for the environment, e.g. ci or prod"
}

variable "signing_keyring_location" {
  description = "The GCP location to create the signing keyring"
}

variable "tf_state_location" {
  description = "The GCP location to store Terraform remote state"
}

variable "hab_keylength" {
  description = "HAB CA RSA key length"
  type        = number
  // From https://github.com/usbarmory/crucible/blob/master/hab/const.go#L13
  default = 2048
}

variable "hab_num_intermediates" {
  description = "Number of HAB intermediate CAs"
  default     = 4
}

variable "hab_pki_lifetime" {
  description = "Lifetime for HAB PKI certs in seconds"
  default     = 788400000 // 25 years
}

variable "hab_ci_revision" {
  description = "Revision count for CI HAB PKI certs. This must be incremented if these certs are regenerated for any reason"
}