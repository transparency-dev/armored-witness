variable "env" {
  description = "Unique identifier for the env, e.g. ci or prod"
  type        = string
}

variable "log_shard" {
  type        = number
  description = "The shard number of the active log. Also corresponds to the KMS crypto key version to use."
}

variable "origin_prefix" {
  type        = string
  description = "Identifier for the log identity. Will be concatenated with the log_shard for the full name."
}

variable "tamago_version" {
  type        = string
  description = "TamaGo version to compile armored-witness firmware with"
}

variable "recovery_tamago_version" {
  type        = string
  description = "TamaGo version to compile recovery image with"
  default     = "1.22.6" # Pin to this by default since armory-ums is not yet compatible with 1.23.x
}

variable "armory_ums_version" {
  type        = string
  description = "Full git commit hash for the armory-ums repo to use when building the recovery image"
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
  type        = string
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

variable "firmware_base_url" {
  description = "Base URL used to construct the log and firmware bucket URLs"
  type        = string
}

variable "rest_distributor_base_url" {
  description = "Base URL for the checkpoint distributor"
  type        = string
}

variable "bastion_addr" {
  description = "Host:port of the bastion server"
  type        = string
  default     = ""
}

