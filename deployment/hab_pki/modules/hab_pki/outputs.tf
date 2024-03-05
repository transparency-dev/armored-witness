output "hab_img_id" {
  description = "HAB IMG cert for the first SRK intermediate"
  value       = google_privateca_certificate.hab_img[1].id
}

output "hab_img_key" {
  description = "HAB IMG key"
  value       = data.google_kms_crypto_key_version.hab_img[1].name
}

output "hab_csf_id" {
  description = "HAB CSF cert for the first SRK intermediate"
  value       = google_privateca_certificate.hab_csf[1].id
}

output "hab_csf_key" {
  description = "HAB CSF key"
  value       = data.google_kms_crypto_key_version.hab_csf[1].name
}

output "hab_srk_ca_ids" {
  description = "HAB SRK intermediates"
  value = {
    for k, hab_srk in google_privateca_certificate_authority.hab_srk : k => hab_srk.id
  }
}
