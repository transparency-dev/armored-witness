# Custodians

Welcome, and *Thank you* for volunteering to be a custodian of an ArmoredWitness device!

This page serves as an introduction and starting point for the care and feeding of the devices. If you have any
questions at all which aren't covered here, please don't hesitate to find us on the
[transparency-dev slack](https://transparency-dev.slack.com/)
[[invite]](https://join.slack.com/t/transparency-dev/shared_invite/zt-27pkqo21d-okUFhur7YZ0rFoJVIOPznQ)!

## First things first

You've just received your ArmoredWitness device, and you want to plug it into your network; this is quite natural - it's very exiting!

**However, before you do**, there are two **very important** things to tell you:

1. ⚠️ **NEVER** ⚠️ provide power to the device via the Power-over-Ethernet *and* USB-C sockets at the same time. \
   This is very likely to result in a broken device (symptom: device boots, but is unable to obtain a network link).
2. You **should** verify that your device hasn't been tampered with on its way to you.\
   Instructions on how to do this are below.

### Verifying the device

Let's make sure that your ArmoredWitness hasn't been tampered with on its way to you.

The [verify](/cmd/verify) tool is used for this, it puts the device into `recovery` mode, and
inspects the firmware stored on the MMC to check that it is unmodified, authentic, and present
in the [firmware transparency](transparency.md) log.

You'll need:

* The ArmoredWitness device
* A Linux computer on which you have `root` priviliges, with:
  * Go 1.22 installed
  * A free USB-A socket to which the device can be connected
  * Internet connectivity
    * Used for fetching the tool, and to allow the tool to contact the firmware transparency logs
* A USB-A to USB-C cable
  * (this is important, it *may* work with some USB-C host ports, but we're aware of a few cases where there were issues due to `USB PD`)
* A small implement (e.g. a very small screwdriver) for sliding the DIP switch visible through the small aperture on the underside of the device. \
  This switch is referred to as the "boot switch" below. \
  Sliding it TOWARD the NETWORK port is the "USB" setting, and AWAY from the NETWORK port is the "MMC" setting. \
  The switch *may* have a pliable translucent orange covering over it, this shouldn't stop you from being able to slide the switch.

First, build the `verify` tool:

```bash
GOBIN=${PWD} go install github.com/transparency-dev/armored-witness/cmd/verify@main
```

Then, you'll need to run it as root and follow the prompts (these looks like "🔷🔷🔷 🙋 OPERATOR: please *do something* 🙏").

> Why root? The `verify` tool needs access to `/dev/hidraw` and `/dev/disk/by-id/usb-F-Secure_USB_*` device files
> in order to put the ArmoredWitness into recovery mode, and read from its internal MMC storage.
> If you have the knowledge to do so, you could use `udev` to grant read-write access to `hid` devices, and `read-only`
> access to the block device.

```bash
$ sudo ${PWD}/verify --template=prod
I0412 10:40:28.663780 2193170 main.go:98] Using template flag setting --binaries_url=https://api.transparency.dev/armored-witness-firmware/prod/artefacts/1/
I0412 10:40:28.663867 2193170 main.go:98] Using template flag setting --firmware_log_url=https://api.transparency.dev/armored-witness-firmware/prod/log/1/
I0412 10:40:28.663882 2193170 main.go:98] Using template flag setting --firmware_log_origin=transparency.dev/armored-witness/firmware_transparency/prod/1
I0412 10:40:28.663893 2193170 main.go:98] Using template flag setting --firmware_log_verifier=transparency.dev-aw-ftlog-prod-1+3e6d87ee+Aa3qdhefd2cc/98jV3blslJT2L+iFR8WKHeGcgFmyjnt
I0412 10:40:28.663914 2193170 main.go:98] Using template flag setting --applet_verifier=transparency.dev-aw-applet-prod+d45f2a0d+AZSnFa8GxH+jHV6ahELk6peqVObbPKrYAdYyMjrzNF35
I0412 10:40:28.663927 2193170 main.go:98] Using template flag setting --boot_verifier=transparency.dev-aw-boot-prod+2fa9168e+AR+KIx++GIlMBICxLkf4ZUK5RDlvJuiYUboqX5//RmUm
I0412 10:40:28.663940 2193170 main.go:98] Using template flag setting --recovery_verifier=transparency.dev-aw-recovery-prod+f3710baa+ATu+HMUuO8ZsgaNwP97XMcb/+Ve8W1u1KdFQHNzOyLxx
I0412 10:40:28.663952 2193170 main.go:98] Using template flag setting --hab_target=prod
I0412 10:40:28.663963 2193170 main.go:98] Using template flag setting --os_verifier_1=transparency.dev-aw-os1-prod+985bdfd2+AV7mmRamQp6VC9CutzSXzqtNhYNyNmQQRcLX07F6qlC1
I0412 10:40:28.663982 2193170 main.go:98] Using template flag setting --os_verifier_2=transparency.dev-aw-os2-prod+662add8c+AebLJIKJhx57T3mWmHKe0sasFnXmtIQNTGRaoj2PQLrY
I0412 10:40:29.681000 2193170 fetcher.go:88] Fetching RECOVERY bin from "8271e2a8ccefb6c4df48889fcbb35343511501e3bcd527317d9e63e2ac7349e3"
I0412 10:40:29.879505 2193170 main.go:217] Successfully fetched and verified recovery image
I0412 10:40:29.879519 2193170 main.go:218] ----------------------------------------------------------------------------------------------
I0412 10:40:29.879523 2193170 main.go:219] 🔷🔷🔷 🙋 OPERATOR: please ensure boot switch is set to USB, and then connect device 🙏
I0412 10:40:29.879526 2193170 main.go:220] ----------------------------------------------------------------------------------------------
I0412 10:40:29.879530 2193170 main.go:223] Recovery firmware is 1924096 bytes + 16384 bytes HAB signature
I0412 10:40:29.879540 2193170 recovery.go:64] Waiting for device to be detected...
I0412 10:46:13.033524 2193170 sdp.go:85] found device 15a2:007d Freescale SemiConductor Inc  SE Blank 6UL
I0412 10:46:13.092825 2193170 sdp.go:111] Attempting to SDP boot device /dev/hidraw0
I0412 10:46:13.092912 2193170 sdp.go:123] Loading DCD at 0x00910000 (976 bytes)
I0412 10:46:13.096387 2193170 sdp.go:128] Loading imx to 0x8000f400 (1940480 bytes)
I0412 10:46:14.288277 2193170 sdp.go:133] Sending jump address to 0x8000f400
I0412 10:46:14.288651 2193170 sdp.go:138] Serial download on /dev/hidraw0 complete
I0412 10:46:15.289152 2193170 recovery.go:51] Witness device booting recovery image
I0412 10:46:15.289210 2193170 recovery.go:106] Waiting for block device to appear
I0412 10:46:18.876369 2193170 recovery.go:118] Waiting for block device to settle...
I0412 10:46:19.897030 2193170 main.go:230] ✅ Detected device "/dev/hidraw0"
I0412 10:46:19.897079 2193170 main.go:231] ✅ Detected blockdevice /dev/disk/by-id/usb-F-Secure_USB_armory_Mk_II_720A9DEAD4413E39-0:0
I0412 10:46:19.900369 2193170 main.go:370] Found config at block 0x4fb0
I0412 10:46:19.900394 2193170 main.go:375] Reading 0x2d6c00 bytes of firmware from MMC byte offset 0x400
I0412 10:46:20.045080 2193170 main.go:370] Found config at block 0x5000
I0412 10:46:20.045122 2193170 main.go:375] Reading 0xdcfe65 bytes of firmware from MMC byte offset 0xa0a000
I0412 10:46:20.765940 2193170 main.go:370] Found config at block 0x200000
I0412 10:46:20.765988 2193170 main.go:375] Reading 0xf09521 bytes of firmware from MMC byte offset 0x4000a000
I0412 10:46:21.695851 2193170 main.go:291]   ✅ Bootloader: proof bundle is self-consistent
I0412 10:46:21.695923 2193170 main.go:314]   ✅ Bootloader: proof bundle checkpoint(@7) is consistent with current view of log(@7)
I0412 10:46:21.714944 2193170 main.go:291]   ✅ TrustedOS: proof bundle is self-consistent
I0412 10:46:21.715029 2193170 main.go:314]   ✅ TrustedOS: proof bundle checkpoint(@7) is consistent with current view of log(@7)
I0412 10:46:21.735579 2193170 main.go:291]   ✅ TrustedApplet: proof bundle is self-consistent
I0412 10:46:21.735666 2193170 main.go:314]   ✅ TrustedApplet: proof bundle checkpoint(@7) is consistent with current view of log(@7)
I0412 10:46:21.735672 2193170 main.go:128] ✅ Device verified OK!
I0412 10:46:21.735681 2193170 main.go:129] ----------------------------------------------------------------------------------------------
I0412 10:46:21.735685 2193170 main.go:130] 🔷🔷🔷 🙋 OPERATOR: please ensure boot switch is set to MMC, and then reboot device 🙏
I0412 10:46:21.735689 2193170 main.go:131] ---------------------------------------------------------------------------------------------- 
```

The block of ✅ green ticks towards the end indicates that the firmware on the device was successfully verified.

If you see these green ticks when running against your device, then you can be sure that:

1. Your ArmoredWitness has not been tampered with on its journey from us to you.
1. All 3 pieces of firmware on the device are publicly discoverable via the transparency log.
   * This enables anyone running the [verify_build](/cmd/verify_build) tool to check that all the firmware in the
     log is reproducibly built, and, consequently, to inspect the source code for all released versions of the firmware to see that it's doing only what it claims.

If you see ❌ **red crosses** here this means something is wrong and the tool was unable to verify the device.
You should unplug the device and contact us, perferably via a publicly-visible channel such as the
[transparency-dev slack](https://transparency-dev.slack.com/)
[[invite]](https://join.slack.com/t/transparency-dev/shared_invite/zt-27pkqo21d-okUFhur7YZ0rFoJVIOPznQ).

## Normal operation

For normal operation, the device will need:

* Power (~1.5 Watts normally)\
  Either from:
  * PoE if your switch supports this (or you are using a PoE injector), or
  * Via a USB-A to USB-C cable
  Note that **PoE MUST NOT be used when powering over USB**
* Internet access\
  network configuration should be provided via DHCP.

Ensure that, having completed the verify step above, the sliding switch on the bottom of the device has been
slid fully back towards the USB end of the board, and connect the device to the network (and USB for power,
if not using PoE).

### LEDs

The device has four LEDs on board:

* Green and orange, on the RJ45 network connector.
* White and blue, on the top-side of the board, visible through lenses on the enclosure.

The normal states for the white and blue LEDs are listed in the table below:

| LED    | State                           | Meaning
|---------------|--------------------------|--------------------------------------
| Orange        | On                       | No link present on network cable
| Green         | On                       | Good link detected on network cable
| Green         | Blinking                 | Network transmitter/receiver active
| White         | Blinking once per second | The Secure Monitor ("OS") is running ok
| White         | On / off for > 5s        | The Secure Monitor ("OS") has crashed
| Blue          | Blinking once per second | The witness applet is running ok
| Blue          | On / off for > 5s        | The witness applet has crashed/exited
| White & Blue  | Dimly lit                | The device is unable to boot (see below)

## Troubleshooting

### Blue & white LEDs are dim, and device doesn't seem to be doing anything

This indicates that the device was not able to boot, there are two possible causes:

1. The switch on the bottom of the device is in the wrong position.\
   Ensure that the switch is fully positioned towards the USB port, and power-cycle the device.
1. The bootloader firmware is corrupt, or has an incorrect HAB signature.\
   The device will need to be re-provisioned. See instructions [below]{#reinstall}.

### Blue & white LEDs are blinking together/in opposite states

This is fine, this shows that both parts of the firmware are working.

### I've been asked to get logs or other info from the witness

If things seems to have gone awry, we may ask you to collect logs or some other status info from the device
and send it to us.

To do this, you'll generally use the `witnessctl` tool from the github.com/transparency-dev/armored-witness-os repo, and correspondingly will need to have the device plugged in via a USB-A to USB-C cable.
**Note that if you are powering the device via PoE, you MUST unplug the network cable first**.

First build the binary:

```bash
GOBIN=${PWD} go install github.com/transparency-dev/armored-witness-os/cmd/witnessctl@main
```

Then you'll need to run it, with flags as below, using sudo/as root in order for the tool to be
able to read and write to the `/dev/hidraw` devices so as to communicate with the ArmoredWitness OS.

#### Status

Fetch the ArmoredWitness device's status with `witnessctl -s`, e.g.:

```bash
$ sudo ${PWD}/witnessctl -s
👁️‍🗨️ @ /dev/hidraw0
----------------------------------------------------------- Trusted OS ----
Serial number ..............: CA6B65D9D4992516
Secure Boot ................: true
SRK hash ...................: b8ba457320663bf006accd3c57e06720e63b21ce5351cb91b4650690bb08d85a
Revision ...................: d009fc3
Version ....................: 0.3.1710844364-incompatible
Runtime ....................: go1.22.0 tamago/arm
Link .......................: false
IdentityCounter ............: 0
Witness/Identity ...........: ArmoredWitness-wispy-snow+834b82b1+ASho7B13t7PhXoLr43ppVFCHEpTSajIybNRYMSi8XR1Q
Witness/IP .................: 10.0.22.57
Witness/AttestationKey .....: AW-ID-Attestation-CA6B65D9D4992516+7eb8d369+ATWcbyKw4qQ+8s7WPwdaDpSB3RlDFw9Ja+d48z5Qsjx2
```

#### Logs (console & crash)

##### Live console

To get the console ("live") logs from the running device, use `witnessctl -l`, e.g:

```bash
❯ sudo ${PWD}/witnessctl -l
👁️‍🗨️ @ /dev/hidraw0
00:00:03 tamago/arm (go1.22.0) • TEE security monitor (Secure World system/monitor) • d009fc3
00:00:03 SM version verification (0.3.1710844364-incompatible)
00:00:03 RPMB program key flag already fused
00:00:03 Loaded OS from slot B
00:00:03 SM log verification pub: transparency.dev-aw-ftlog-ci-3+3f689522+Aa1Eifq6rRC8qiK+bya07yV1fXyP156pEMsX7CFBC6gg
00:00:03 SM applet verification pub: transparency.dev-aw-applet-ci+3ff32e2c+AV1fgxtByjXuPjPfi0/7qTbEBlPGGCyxqr6ZlppoLOz3
00:00:04 Loaded applet from slot A
00:00:04 SM Verifying applet bundle
00:00:06 SM Loaded applet version 0.3.1710504952-incompatible
00:00:07 SM applet loaded addr:0x20000000 entry:0x200876dc size:16963776
00:00:07 SM applet started mode:USR sp:0x30000000 pc:0x200876dc ns:false
00:00:07 SM IRQ handling error: unexpected IRQ 1023
I0101 00:00:07.770420       3 main.go:126] tamago/arm (go1.22.0) • TEE user applet • 488ede8
00:00:07 SM starting network
I0101 00:00:07.804469       3 main.go:173] ----------------------------------------------------------- Trusted OS ----
I0101 00:00:07.813636       3 main.go:173] Serial number ..............: CA6B65D9D4992516
I0101 00:00:07.821390       3 main.go:173] Secure Boot ................: true
I0101 00:00:07.828166       3 main.go:173] SRK hash ...................: b8ba457320663bf006accd3c57e06720e63b21ce5351cb91b4650690bb08d85a
I0101 00:00:07.840175       3 main.go:173] Revision ...................: d009fc3
I0101 00:00:07.847210       3 main.go:173] Version ....................: 0.3.1710844364-incompatible
I0101 00:00:07.856029       3 main.go:173] Runtime ....................: go1.22.0 tamago/arm
I0101 00:00:07.864089       3 main.go:173] Link .......................: true
I0101 00:00:07.870880       3 main.go:173] IdentityCounter ............: 0
I0101 00:00:07.877396       3 main.go:173] Witness ....................: <no status>
I0101 00:00:08.174043       3 main.go:188] Attestation key:
AW-ID-Attestation-CA6B65D9D4992516+7eb8d369+ATWcbyKw4qQ+8s7WPwdaDpSB3RlDFw9Ja+d48z5Qsjx2
I0101 00:00:08.185213       3 main.go:189] Attested identity key:
ArmoredWitness ID attestation v1
CA6B65D9D4992516
0
ArmoredWitness-wispy-snow+834b82b1+ASho7B13t7PhXoLr43ppVFCHEpTSajIybNRYMSi8XR1Q

— AW-ID-Attestation-CA6B65D9D4992516 frjTabuAin+geejJ3AShNmeyaivFrv7J8hL+eNqJkp2YgZGQ5jZStLLTWUl/orretpYtWg5pgcgzJF2lLRIqR/bpDAg=
I0101 00:00:08.229161       3 main.go:209] Opening storage...
I0101 00:00:08.237440       3 main.go:411] CardInfo: {MMC:true SD:false HC:true HS:true DDR:false Rate:150 BlockSize:512 Blocks:30576640 CID:[184 160 140 178 241 16 65 51 52 65 56 53 3 1 214 0]}
I0101 00:00:08.262174       3 main.go:211] Storage opened.
00:00:08 SM registering applet event handler g:0x20dd0fc8 p:0x20c2e008
I0101 00:00:08.313678       3 net.go:189] Starting DHCPClient...
I0101 00:00:08.327403       3 net.go:120] DHCPC: lease update - old: /0, new: 10.0.22.57/22
I0101 00:00:08.334072       3 net.go:151] DHCPC: Acquired 10.0.22.57/22
I0101 00:00:08.340382       3 net.go:203] DHCPC: Using DNS server(s) [10.0.20.1:53]
I0101 00:00:08.347508       3 net.go:215] DHCPC: Using Gateway 10.0.20.1
I0101 00:00:08.354104       3 net.go:171] DHCP: running f
I0101 00:00:08.359188       3 main.go:275] TA Version:0.3.1710504952-incompatible MAC:da:47:2c:51:e8:70 IP:10.0.22.57/22 GW:[10.0.20.0/22 nic 1 0.0.0.0/0 via 10.0.20.1 nic 1] DNS:[10.0.20.1:53]
00:00:13 Grabbing log messages...
```

##### Crash logs

The "Crash" logs are stored if the device detects a problem and needs to restart the witness.

To fetch these logs, use the `witnessctl -L` command:

```bash
sudo ${PWD}/witnessctl -L
👁️‍🗨️ @ /dev/hidraw0
...
```

An empty response means there are no crash logs.

Be aware that there can be a relatively large amount of log returned, it might be wise to redirect (or `tee`) the
output from the above command into a file.

Fetching the crash log from the device does not erase it, so you can always re-run the above command.

Note that the device only stores logs from the most recent witness restart; a previous crash log will get
overwritten in the event of another unexpected restart.

## Reinstall

You will need to re-run the provision tool, for now, it's probably best to get in touch with us via the
[transparency-dev slack](https://transparency-dev.slack.com/)
[[invite]](https://join.slack.com/t/transparency-dev/shared_invite/zt-27pkqo21d-okUFhur7YZ0rFoJVIOPznQ)
and we'll walk you through it - it's not hard to use the tool, BUT we would certainly like to know that it
seems to be necessary to do it.

## More info

