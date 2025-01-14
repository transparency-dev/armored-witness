variable "project_id" {
  type        = number
  description = "The project ID to host the cluster in"
}

variable "project_name" {
  type        = string
  description = "The string project ID"
}

variable "signing_keyring_location" {
  type        = string
  description = "The GCP location to create the signing keyring"
}

variable "tf_state_location" {
  type        = string
  description = "The GCP location to store Terraform remote state"
}

variable "serve_domain" {
  description = "The domain we'll use to serve"
  type        = string
}

variable "tls" {
  description = "Enable TLS"
  type        = bool
}

variable "distributor_prod_host" {
  type        = string
  description = "Host name serving distributor service API (prod)"
}
variable "distributor_prod_port" {
  description = "Port on distributor_host where distributor service API is served (prod)"
  type        = number
}
variable "distributor_ci_host" {
  type        = string
  description = "Host name serving distributor service API (ci)"
}
variable "distributor_ci_port" {
  description = "Port on distributor_host where distributor service API is served (ci)"
  type        = number
}
variable "distributor_dev_host" {
  type        = string
  description = "Host name serving distributor service API (dev)"
}
variable "distributor_dev_port" {
  description = "Port on distributor_host where distributor service API is served (dev)"
  type        = number
}

variable "lb_name" {
  type        = string
  description = "Name of the load balancer"
}
