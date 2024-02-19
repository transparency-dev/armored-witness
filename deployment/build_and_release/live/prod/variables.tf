variable "project_id" {
  description = "The project ID to host the cluster in"
}

variable "env" {
  description = "Unique identifier for the env, e.g. ci or prod"
  type        = string
}

variable "bucket_env" {
  description = "Identifier for the env to use for existing buckets, e.g. '-ci' or ''. TODO: migrate/rename these resources."
  type        = string
}

variable "signing_keyring_location" {
  description = "The GCP location to create the signing keyring"
}

variable "tf_state_location" {
  description = "The GCP location to store Terraform remote state"
}
