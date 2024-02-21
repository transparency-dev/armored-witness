variable "env" {
  description = "Unique identifier for the env, e.g. ci or prod"
  type        = string
}

variable "cloudbuild_path" {
  description = "The path of the Cloud Build config in the Github repo"
  type        = string
}

