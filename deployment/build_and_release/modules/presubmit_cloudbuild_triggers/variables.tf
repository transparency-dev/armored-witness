variable "env" {
  description = "Unique identifier for the env, e.g. ci or prod"
  type        = string
}

variable "build_components" {
  type = map(object({
    repo            = string
    cloudbuild_path = string
  }))
}

