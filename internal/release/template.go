package release

const (
	templateCI   = "ci"
	templateProd = "prod"
)

var (
	Templates = map[string]map[string]string{
		templateCI: {
			"binaries_url":          "https://api.transparency.dev/armored-witness-firmware/ci/artefacts/4/",
			"firmware_log_url":      "https://api.transparency.dev/armored-witness-firmware/ci/log/4/",
			"firmware_log_origin":   "transparency.dev/armored-witness/firmware_transparency/ci/4",
			"firmware_log_verifier": "transparency.dev-aw-ftlog-ci-4+30fe79e3+AUDoas+smwQDTlYbTzbEcAW+N6WyvB/4CysMWjpnRgat",
			"applet_verifier":       "transparency.dev-aw-applet-ci+3ff32e2c+AV1fgxtByjXuPjPfi0/7qTbEBlPGGCyxqr6ZlppoLOz3",
			"boot_verifier":         "transparency.dev-aw-boot-ci+9f62b6ac+AbnipFmpRltfRiS9JCxLUcAZsbeH4noBOJXbVD3H5Eg4",
			"recovery_verifier":     "transparency.dev-aw-recovery-ci+cc699423+AarlJMSl0rbTMf31B5o9bqc6PHorwvF1GbwyJRXArbfg",
			"os_verifier_1":         "transparency.dev-aw-os1-ci+7a0eaef3+AcsqvmrcKIbs21H2Bm2fWb6oFWn/9MmLGNc6NLJty2eQ",
			"os_verifier_2":         "transparency.dev-aw-os2-ci+af8e4114+AbBJk5MgxRB+68KhGojhUdSt1ts5GAdRIT1Eq9zEkgQh",
			"hab_target":            "ci",
		},
		templateProd: {
			"binaries_url":          "https://api.transparency.dev/armored-witness-firmware/prod/artefacts/1/",
			"firmware_log_url":      "https://api.transparency.dev/armored-witness-firmware/prod/log/1/",
			"firmware_log_origin":   "transparency.dev/armored-witness/firmware_transparency/prod/1",
			"firmware_log_verifier": "transparency.dev-aw-ftlog-prod-1+3e6d87ee+Aa3qdhefd2cc/98jV3blslJT2L+iFR8WKHeGcgFmyjnt",
			"applet_verifier":       "transparency.dev-aw-applet-prod+d45f2a0d+AZSnFa8GxH+jHV6ahELk6peqVObbPKrYAdYyMjrzNF35",
			"boot_verifier":         "transparency.dev-aw-boot-prod+2fa9168e+AR+KIx++GIlMBICxLkf4ZUK5RDlvJuiYUboqX5//RmUm",
			"recovery_verifier":     "transparency.dev-aw-recovery-prod+f3710baa+ATu+HMUuO8ZsgaNwP97XMcb/+Ve8W1u1KdFQHNzOyLxx",
			"os_verifier_1":         "transparency.dev-aw-os1-prod+985bdfd2+AV7mmRamQp6VC9CutzSXzqtNhYNyNmQQRcLX07F6qlC1",
			"os_verifier_2":         "transparency.dev-aw-os2-prod+662add8c+AebLJIKJhx57T3mWmHKe0sasFnXmtIQNTGRaoj2PQLrY",
			"hab_target":            "prod",
		},
	}
)
