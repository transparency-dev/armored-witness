# Device Firmware Verification Tool

The `verify` command is a tool which will check that the firmware currently installed
on an ArmoredWitness device is authentic and discoverable via firmware transparency.

There are 3 types of firmware installed on an ArmoredWitness device:

1. Bootloader
2. Secure Monitor ("OS")
3. Witness applet

Each of these is stored alongside a `ProofBundle` which contains firmware transparency
information about the firmware image itself.

Broadly, the tool performs the following operations:

1. Puts the device into recovery mode, where its MMC storage becomes visible as a USB Mass Storage device.
2. Extracts the installed firmware & proof bundles for each of the 3 types of firmware
(bootloader, OS, applet).
3. For each of the firmware types, it:
    * Verifies the signature on the associated `manifest`.
    * Hashes the firmware image, and asserts that it matches the one in the `manifest`.
    * Verifies that the firmware `manifest` is present within the expected firmware transparency log.

For more detailed information about the firmware transparency concepts and metadata, please
see the [firmware transparency](/docs/transparency.md) doc.

## Usage

This tool is only useful if you have a provisioned ArmoredWitness device to hand.

Ensure that the device is completely unplugged, with neither USB nor LAN connected.

Locate the small black switch on the underside of the board, this switch has two positions:

* USB (with the white slider towards the RJ45 socket)
* MMC (with the white slider away from the RJ45 socket)

The `verify` tool will ask you move this switch a couple of times during the process - please
follow the "🙏 Operator, please _______" prompts from the tool. You may find a small pointed
object helpful for moving the slider on the switch, it can be quite fiddly!

Once you've finished verifying the device by running the command below, don't forget to follow the
prompt to return the switch to its original position, and reboot the device (e.g. by unplugging and
plugging it back in)! If you do happen to forget, do not panic, all is well - you'll either see:

* A solid bright blue LED --> the device has not been rebooted yet
    1. Ensure you returned the switch to the `MMC` location,
    2. Restart the device.
* Dim white and blue LEDs --> the device was restarted, but the switch is still in the `USB` location:
    1. Change the switch to the `MMC` location,
    2. Restart the device.

When you're ready, the tool can be run directly from the repo and will need to be told which release
train the device was provisioned into - this is likely to be `prod` if you are a custodian, or
`ci` if you are on the team - adjust the `template` flag before running the command
below:

```bash
# This command needs to either be run as root, or have access to USB-related /dev
# files granted through some other mechanism.
$ go build github.com/transparency-dev/armored-witness/cmd/verify@main
$ sudo ~/go/bin/verify -template=ci
I0311 18:44:50.422040  243805 main.go:96] Using template flag setting --boot_verifier=transparency.dev-aw-boot-ci+9f62b6ac+AbnipFmpRltfRiS9JCxLUcAZsbeH4noBOJXbVD3H5Eg4
I0311 18:44:50.422138  243805 main.go:96] Using template flag setting --recovery_verifier=transparency.dev-aw-recovery-ci+cc699423+AarlJMSl0rbTMf31B5o9bqc6PHorwvF1GbwyJRXArbfg
I0311 18:44:50.422148  243805 main.go:96] Using template flag setting --os_verifier_1=transparency.dev-aw-os1-ci+7a0eaef3+AcsqvmrcKIbs21H2Bm2fWb6oFWn/9MmLGNc6NLJty2eQ
I0311 18:44:50.422157  243805 main.go:96] Using template flag setting --binaries_url=https://api.transparency.dev/armored-witness-firmware/ci/artefacts/2/
I0311 18:44:50.422164  243805 main.go:96] Using template flag setting --applet_verifier=transparency.dev-aw-applet-ci+3ff32e2c+AV1fgxtByjXuPjPfi0/7qTbEBlPGGCyxqr6ZlppoLOz3
I0311 18:44:50.422171  243805 main.go:96] Using template flag setting --firmware_log_verifier=transparency.dev-aw-ftlog-ci-2+f77c6276+AZXqiaARpwF4MoNOxx46kuiIRjrML0PDTm+c7BLaAMt6
I0311 18:44:50.422178  243805 main.go:96] Using template flag setting --os_verifier_2=transparency.dev-aw-os2-ci+af8e4114+AbBJk5MgxRB+68KhGojhUdSt1ts5GAdRIT1Eq9zEkgQh
I0311 18:44:50.422186  243805 main.go:96] Using template flag setting --hab_target=ci
I0311 18:44:50.422195  243805 main.go:96] Using template flag setting --firmware_log_url=https://api.transparency.dev/armored-witness-firmware/ci/log/2/
I0311 18:44:50.422205  243805 main.go:96] Using template flag setting --firmware_log_origin=transparency.dev/armored-witness/firmware_transparency/ci/2
I0311 18:44:55.917161  243805 fetcher.go:88] Fetching RECOVERY bin from "8271e2a8ccefb6c4df48889fcbb35343511501e3bcd527317d9e63e2ac7349e3"
I0311 18:44:56.047071  243805 main.go:218] Successfully fetched and verified recovery image
I0311 18:44:56.047114  243805 main.go:219] ----------------------------------------------------------------------------------------------
I0311 18:44:56.047124  243805 main.go:220] 🙏 Operator, please ensure boot switch is set to USB, and then connect device 🙏
I0311 18:44:56.047132  243805 main.go:221] ----------------------------------------------------------------------------------------------
I0311 18:44:56.047144  243805 main.go:224] Recovery firmware is 1924096 bytes + 16384 bytes HAB signature
I0311 18:44:56.047168  243805 recovery.go:63] Waiting for device to be detected...
I0311 18:45:16.109285  243805 sdp.go:85] found device 15a2:007d Freescale SemiConductor Inc  SE Blank 6UL
I0311 18:45:16.168919  243805 sdp.go:111] Attempting to SDP boot device /dev/hidraw0
I0311 18:45:16.169042  243805 sdp.go:123] Loading DCD at 0x00910000 (976 bytes)
I0311 18:45:16.172336  243805 sdp.go:128] Loading imx to 0x8000f400 (1940480 bytes)
I0311 18:45:17.404520  243805 sdp.go:133] Sending jump address to 0x8000f400
I0311 18:45:17.404898  243805 sdp.go:138] Serial download on /dev/hidraw0 complete
I0311 18:45:18.405382  243805 recovery.go:50] Witness device booting recovery image
I0311 18:45:18.405419  243805 recovery.go:105] Waiting for block device to appear
I0311 18:45:21.819356  243805 recovery.go:118] Waiting for block device to settle...
I0311 18:45:26.819676  243805 main.go:231] ✅ Detected device "/dev/hidraw0"
I0311 18:45:26.819713  243805 main.go:232] ✅ Detected blockdevice /dev/disk/by-id/usb-F-Secure_USB_armory_Mk_II_CA6B65D9D4992516-0:0
I0311 18:45:26.825808  243805 main.go:371] Found config at block 0x4fb0
I0311 18:45:26.825835  243805 main.go:376] Reading 0x2d6c00 bytes of firmware from MMC byte offset 0x400
I0311 18:45:26.966716  243805 main.go:371] Found config at block 0x5000
I0311 18:45:26.966954  243805 main.go:376] Reading 0xdbbe24 bytes of firmware from MMC byte offset 0xa0a000
I0311 18:45:27.633841  243805 main.go:371] Found config at block 0x200000
I0311 18:45:27.633871  243805 main.go:376] Reading 0x102cee5 bytes of firmware from MMC byte offset 0x4000a000
I0311 18:45:28.441575  243805 main.go:292]   ✅ Bootloader: proof bundle is self-consistent
I0311 18:45:28.449098  243805 main.go:315]   ✅ Bootloader: proof bundle checkpoint(@42) is consistent with current view of log(@50)
I0311 18:45:28.482859  243805 main.go:292]   ✅ TrustedOS: proof bundle is self-consistent
I0311 18:45:28.482947  243805 main.go:315]   ✅ TrustedOS: proof bundle checkpoint(@49) is consistent with current view of log(@50)
I0311 18:45:28.505108  243805 main.go:292]   ✅ TrustedApplet: proof bundle is self-consistent
I0311 18:45:28.505194  243805 main.go:315]   ✅ TrustedApplet: proof bundle checkpoint(@50) is consistent with current view of log(@50)
I0311 18:45:28.505208  243805 main.go:126] ✅ Device verified OK!
I0311 18:45:28.505212  243805 main.go:127] ----------------------------------------------------------------------------------------------
I0311 18:45:28.505215  243805 main.go:128] 🙏 Operator, please ensure boot switch is set to MMC, and then reboot device 🙏
I0311 18:45:28.505222  243805 main.go:129] ----------------------------------------------------------------------------------------------
```

In the above run, we can see a successfully verified device which was provisioned onto the `ci` release train.

## Digging deeper

If you are curious or want to dig further into the firmware transparency artefacts and verification, you can add a `-v=1` flag to
see firmware `config` and `manifest` structures too. e.g:

```
...
I0311 18:58:46.884302  244428 main.go:371] Found config at block 0x200000
I0311 18:58:46.884413  244428 main.go:374] Config:
{
  "Offset": 1073782784,
  "Size": 16961253,
  "Signatures": null,
  "Bundle": {
    "Checkpoint": "dHJhbnNwYXJlbmN5LmRldi9hcm1vcmVkLXdpdG5lc3MvZmlybXdhcmVfdHJhbnNwYXJlbmN5L2NpLzIKNTAKUHpyZmV6MGovclc5Z2xNYWlsQ2dHaFJpcGYzVHBpZnFuUUxZeGZKdXljaz0KCuKAlCB0cmFuc3BhcmVuY3kuZGV2LWF3LWZ0bG9nLWNpLTIgOTN4aWRtcmpFbDFqZ1NJMlVLYnF3S3o2M3J1MjRnUGlWMVRTRi81TmYyc3VxWmtLQzNBWmUzTVg0TzJ0OEh6U1dXRStMS1hZMDFmQjNWbmw1NnBpa1VwL2xRaz0K",
    "Manifest": "ewogICJzY2hlbWFfdmVyc2lvbiI6IDAsCiAgImNvbXBvbmVudCI6ICJUUlVTVEVEX0FQUExFVCIsCiAgImdpdCI6IHsKICAgICJ0YWdfbmFtZSI6ICIwLjMuMTcxMDE3ODI2OC1pbmNvbXBhdGlibGUiLAogICAgImNvbW1pdF9maW5nZXJwcmludCI6ICI5YjI4YjMwNDYxY2IzMmRhZDg0YTczNmQzYzRjNzJiMDI0MTAwNjNlIgogIH0sCiAgImJ1aWxkIjogewogICAgInRhbWFnb192ZXJzaW9uIjogIjEuMjIuMCIsCiAgICAiZW52cyI6IFsKICAgICAgIkZUX0xPR19VUkw9aHR0cHM6Ly9hcGkudHJhbnNwYXJlbmN5LmRldi9hcm1vcmVkLXdpdG5lc3MtZmlybXdhcmUvY2kvbG9nLzIiLAogICAgICAiRlRfQklOX1VSTD1odHRwczovL2FwaS50cmFuc3BhcmVuY3kuZGV2L2FybW9yZWQtd2l0bmVzcy1maXJtd2FyZS9jaS9hcnRlZmFjdHMvMiIsCiAgICAgICJMT0dfT1JJR0lOPXRyYW5zcGFyZW5jeS5kZXYvYXJtb3JlZC13aXRuZXNzL2Zpcm13YXJlX3RyYW5zcGFyZW5jeS9jaS8yIiwKICAgICAgIkxPR19QVUJMSUNfS0VZPXRyYW5zcGFyZW5jeS5kZXYtYXctZnRsb2ctY2ktMitmNzdjNjI3NitBWlhxaWFBUnB3RjRNb05PeHg0Nmt1aUlSanJNTDBQRFRtK2M3QkxhQU10NiIsCiAgICAgICJBUFBMRVRfUFVCTElDX0tFWT10cmFuc3BhcmVuY3kuZGV2LWF3LWFwcGxldC1jaSszZmYzMmUyYytBVjFmZ3h0QnlqWHVQalBmaTAvN3FUYkVCbFBHR0N5eHFyNlpscHBvTE96MyIsCiAgICAgICJPU19QVUJMSUNfS0VZMT10cmFuc3BhcmVuY3kuZGV2LWF3LW9zMS1jaSs3YTBlYWVmMytBY3Nxdm1yY0tJYnMyMUgyQm0yZldiNm9GV24vOU1tTEdOYzZOTEp0eTJlUSIsCiAgICAgICJPU19QVUJMSUNfS0VZMj10cmFuc3BhcmVuY3kuZGV2LWF3LW9zMi1jaSthZjhlNDExNCtBYkJKazVNZ3hSQis2OEtoR29qaFVkU3QxdHM1R0FkUklUMUVxOXpFa2dRaCIsCiAgICAgICJSRVNUX0RJU1RSSUJVVE9SX0JBU0VfVVJMPWh0dHBzOi8vYXBpLnRyYW5zcGFyZW5jeS5kZXYvY2kiLAogICAgICAiQkVFPTEiLAogICAgICAiREVCVUc9MSIsCiAgICAgICJTUktfSEFTSD1iOGJhNDU3MzIwNjYzYmYwMDZhY2NkM2M1N2UwNjcyMGU2M2IyMWNlNTM1MWNiOTFiNDY1MDY5MGJiMDhkODVhIgogICAgXQogIH0sCiAgIm91dHB1dCI6IHsKICAgICJmaXJtd2FyZV9kaWdlc3Rfc2hhMjU2IjogIlBXb3FoYVBUSk93Qk5GcnV3NlBuSm1vR0pydXBxRzNVU2FwL0VCUmhrNWc9IgogIH0KfQoK4oCUIHRyYW5zcGFyZW5jeS5kZXYtYXctYXBwbGV0LWNpIFAvTXVMTFZKSk1pMXBSNkdXVlZJQTZJZDArTnQzYUFJNFRZZ3dHUlY4TFBVbDlOVlhSZzJXVGVMaTBHanI3aWF6S0dtdmpxNDFqd0F0ZUZOVUR6ZkpKR0cwZ289Cg==",
    "LogIndex": 49,
    "InclusionProof": [
      "qZjJgYmVs+Mv+RsL1oZ82OS0uU11JTZ7jogzZnUgtr4=",
      "jDgUAXfkstK/nxCgfFZD1J70fHmQ+0wDlzx7iLVpK8w=",
      "E7AE6DLeB0wuEIRTczjIMBR8/0uCccTdRuCWB7oWcqw="
    ]
  }
}
...
I0311 18:58:47.742581  244428 main.go:285] TrustedOS extracted firmware has base64 hash: soHj8D1IqFWG2PAKC3Hzmd2cBf8saXPHZQLUksAvGO8=
I0311 18:58:47.742592  244428 main.go:286] TrustedOS Manifest:
{
  "schema_version": 0,
  "component": "TRUSTED_OS",
  "git": {
    "tag_name": "0.3.1709923158-incompatible",
    "commit_fingerprint": "b8ef0168f62f72083f40e222afb168f9ba8d272d"
  },
  "build": {
    "tamago_version": "1.22.0",
    "envs": [
      "LOG_ORIGIN=transparency.dev/armored-witness/firmware_transparency/ci/2",
      "LOG_PUBLIC_KEY=transparency.dev-aw-ftlog-ci-2+f77c6276+AZXqiaARpwF4MoNOxx46kuiIRjrML0PDTm+c7BLaAMt6",
      "APPLET_PUBLIC_KEY=transparency.dev-aw-applet-ci+3ff32e2c+AV1fgxtByjXuPjPfi0/7qTbEBlPGGCyxqr6ZlppoLOz3",
      "OS_PUBLIC_KEY1=transparency.dev-aw-os1-ci+7a0eaef3+AcsqvmrcKIbs21H2Bm2fWb6oFWn/9MmLGNc6NLJty2eQ",
      "OS_PUBLIC_KEY2=transparency.dev-aw-os2-ci+af8e4114+AbBJk5MgxRB+68KhGojhUdSt1ts5GAdRIT1Eq9zEkgQh",
      "BEE=1",
      "DEBUG=1",
      "SRK_HASH=b8ba457320663bf006accd3c57e06720e63b21ce5351cb91b4650690bb08d85a"
    ]
  },
  "output": {
    "firmware_digest_sha256": "soHj8D1IqFWG2PAKC3Hzmd2cBf8saXPHZQLUksAvGO8="
  }
}

— transparency.dev-aw-os1-ci eg6u82hln15lcg8pvB4ME33EI0nb0kvlIc69YLy3LeoABPLtGEGTFOCkLt50mu7NltcUMSRdoSwELeFRiBNDo0JHNAM=
— transparency.dev-aw-os2-ci r45BFJqAqLjzjNZgcTsdJhy3r3pID+1pjM3a6JdfSyGqGpIPnt9aCawTWNig7/71PTOVFn2e3qmYGL8mkOEJ5ODEdQk=
I0311 18:58:47.753047  244428 main.go:292]   ✅ TrustedOS: proof bundle is self-consistent
I0311 18:58:47.753154  244428 main.go:315]   ✅ TrustedOS: proof bundle checkpoint(@49) is consistent with current view of log(@50)
...
```
