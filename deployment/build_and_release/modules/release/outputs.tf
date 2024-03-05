output "firmware_log_buckets" {
  description = "Firmware log GCS bucket names"
  value = {
    for k, bucket in google_storage_bucket.firmware_log : k => bucket.name
  }
}

output "firmware_artefact_buckets" {
  description = "Firmware artefact GCS bucket names"
  value = {
    for k, bucket in google_storage_bucket.firmware : k => bucket.name
  }
}