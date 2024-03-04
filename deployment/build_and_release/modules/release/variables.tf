variable "project_id" {
  description = "The project ID to host the cluster in"
  type        = string
}

variable "env" {
  description = "Unique identifier for the env, e.g. ci or prod"
  type        = string
}

variable "bucket_count" {
  description = "The number of log and firmware buckets to create (each)."
  type        = number
}

variable "signing_keyring_location" {
  description = "The GCP location to create the signing keyring"
  type        = string
}

variable "build_components" {
  type = map(object({
    repo            = string
    cloudbuild_path = string
  }))
}

variable "cloudbuild_trigger_tag" {
  description = <<EOH
    Specifies how the build will be triggered. Exactly one of cloudbuild_trigger_branch or cloudbuild_trigger_tag should be specified.
    See more: https://cloud.google.com/build/docs/api/reference/rest/v1/projects.locations.triggers#BuildTrigger.PushFilter.
  EOH
  type        = string
  default     = ""
}

variable "cloudbuild_trigger_branch" {
  description = <<EOH
    Specifies how the build will be triggered. Exactly one of cloudbuild_trigger_branch or cloudbuild_trigger_tag should be specified.
    See more: https://cloud.google.com/build/docs/api/reference/rest/v1/projects.locations.triggers#BuildTrigger.PushFilter.
  EOH
  type        = string
  default     = ""
}

variable "log_shard" {
  type        = number
  description = "The shard number of the active log. Also corresponds to the KMS crypto key version to use."
}

variable "origin_prefix" {
  type        = string
  description = "Identifier for the log identity. Will be concatenated with the log_shard for the full name."
}

variable "log_name_prefix" {
  type        = string
  description = "The GCS bucket name for the transparency log. Will be concatenated with the log_shard for the full name."
}

variable "firmware_bucket_prefix" {
  type        = string
  description = "The GCS bucket name for the firmware. Will be concatenated with the log_shard for the full name."
}

variable "tamago_version" {
  type        = string
  description = "TamaGo version to compile with"
}

variable "armory_ums_version" {
  type        = string
  description = "Full git commit hash for the armory-ums repo to use when building the recovery image"
}

variable "entries_dir" {
  type        = string
  description = "Specifies where the to-be-sequenced entries are"
}

variable "log_public_key" {
  type        = string
  description = <<-EOT
    [Note verifier string](https://pkg.go.dev/golang.org/x/mod/sumdb/note#hdr-Verifying_Notes)
    for the log
  EOT
}

variable "applet_public_key" {
  type        = string
  description = <<-EOT
    [Note verifier string](https://pkg.go.dev/golang.org/x/mod/sumdb/note#hdr-Verifying_Notes)
    for the applet
  EOT
}

variable "os_public_key1" {
  type        = string
  description = <<-EOT
    First [note verifier string](https://pkg.go.dev/golang.org/x/mod/sumdb/note#hdr-Verifying_Notes)
    for the OS
  EOT
}

variable "os_public_key2" {
  type        = string
  description = <<-EOT
    Second [Note verifier string](https://pkg.go.dev/golang.org/x/mod/sumdb/note#hdr-Verifying_Notes)
    for the OS
  EOT
}

variable "console" {
  description = "If set to `on`, then the bootloader firmware will emit debug logging"
  type        =  string
  default     = ""
}

variable "bee" {
  type        = string
  description = "If '1', compile with BEE flag"
}

variable "debug" {
  type        = string
  description = "If '1', compile with DEBUG flag"
}

variable "checkpoint_cache" {
  type        = string
  description = "Cache-Control header for checkpoint objects"
}

variable "srk_hash" {
  type        = string
  description = <<-EOT
    Pinned CI SRK hash
    This MUST be identical to the _PINNED_SRK_HASH in
    https://github.com/transparency-dev/armored-witness-boot/blob/main/release/cloudbuild_ci.yaml#L223-L224
    and MUST NOT be changed unless you know very well what you're doing,
    otherwise devices will be bricked!
  EOT
}

variable "hab_key_version" {
  type        = number
  description = "Key version of the keys to sign CSF and IMG payloads"
}

variable "hab_revision" {
  description = "Revision count for HAB PKI certs. This must be incremented if these certs are regenerated for any reason"
}

variable "hab_leaf_minor" {
  description = "Revision count for CSF and IMG certs. This allows us to optionally regenerate these certs, while leaving the SRK ones in place."
  type        = string
  default     = ""
}
