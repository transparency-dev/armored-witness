variable "project_id" {
  description = "The project ID to host the cluster in"
}

variable "signing_keyring_location" {
  description = "The GCP location to create the signing keyring"
}

variable "tf_state_location" {
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

variable "distributor_host" {
  description = "Host name serving distributor service API"
}
variable "distributor_port" {
  description = "Port on distributor_host where distributor service API is served"
  type        = number
}

variable "lb_name" {
  description = "Name of the load balancer"
}
