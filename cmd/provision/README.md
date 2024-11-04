# Provision

The `provision` tool is used to provision blank hardware into
production ArmoredWitness devices.

This tool verifies the initial state of a blank ArmoredWitness
device, and then proceeds to flash the various pieces of firmware
onto the device before setting fuses and retrieving the newly
provisioned witness identity.

Note that because of the level of access this tool requires, you
must either run it via `sudo`, or configure `udev` rules etc. to allow
it to inspect & write to appropriate `/dev` files etc.

## Flashing CI or Prod Builds

A device can be configured with builds from the `ci` and `prod` pipelines without
needing much of the tooling in the advanced section of this doc. The first step
is to install the provision tool:

```shell
go install github.com/transparency-dev/armored-witness/cmd/provision@main
```

After this is installed, the tool can be run with the following command to flash
the latest builds of each of the components onto the device:

```shell
sudo $(which provision) \
  --template=${TEMPLATE} \
  --wipe_witness_state
```

where `TEMPLATE` is one of `ci` or `prod`.

Look out for and follow the *"Operator, please do something 🙏"* requests as the provision
process executes.

To **permanently** lock the device to either the `ci` or `prod` releases, add the `--fuse` flag
to the above command.

## Provisioning Dev Builds

### Prerequisites

One of the first things the tool will attempt to do is examine
the firmware transparency (FT) log for the latest available versions of the
4 ArmoredWitness firmware types:

* `BOOT`: <https://github.com/transparency-dev/armored-witness-boot>
* `OS`: <https://github.com/transparency-dev/armored-witness-os>
* `APPLET`: <https://github.com/transparency-dev/armored-witness-applet>
* `RECOVERY`: <https://github.com/usbarmory/armory-ums>

For this to succeed, particularly if using a local FT log during development,
please ensure that you have built and logged all 4 types of firmware -
instructions on how to do this are in the main READMEs of the
`armored-witness-*` firmware repos (instructions for logging the recovery image
are in the `armored-witness-boot` README).

### Provisioning development builds

You will also need to supply the following info as flags to the tool:

| Flag                      | Value
|---------------------------|-------
| `--firmware_log_url`      | Base URL of the FT log. Use a`file://...` for local dev FT log. Must end in trailing `/`
| `--firmware_log_origin`   | Log Origin string, for dev use `DEV.armoredwitness.transparency.dev/${USER}`
| `--firmware_log_verifier` | Log public key in note verifier format
| `--binaries_url`          | Base URL to fetch binaries, probably adjacent to FT log location. Must end in trailing `/`
| `--applet_verifier`       | Applet firmware public key in note verifier format
| `--boot_verifier`         | Bootloader firmware public key in note verifier format
| `--os_verifier_1`         | OS firmware public key 1 in note verifier format
| `--os_verifier_2`         | OS firmware public key 2 in note verifier format
| `--recovery_verifier`     | Recovery firmware public key in note verifier format

If you're going to be using this tool more than just a few times, it may be
helpful to have [`uhubctl`](https://github.com/mvp/uhubctl) installed, and
have access to a compatible USB hub - this will enable you to power-cycle/reboot
the device without having to keep unplugging it.

### Usage

First, build the tool (commands assume PWD is the root of the repo):

```bash
go build ./cmd/provision
```

Now run with elevated privileges, e.g.:

```bash
$ sudo ./provision \
    --binaries_url=${BINARIES_URL}/ \
    --firmware_log_url=${LOG_URL}/ \
    --firmware_log_origin=${LOG_ORIGIN} \
    --firmware_log_verifier=$(cat ${LOG_PUBLIC_KEY}) \
    --applet_verifier=$(cat ${APPLET_PUBLIC_KEY}) \
    --boot_verifier=$(cat ${BOOT_PUBLIC_KEY}) \
    --recovery_verifier=$(cat ${RECOVERY_PUBLIC_KEY}) \
    --os_verifier_1=$(cat ${OS_PUBLIC_KEY1}) \
    --os_verifier_2=$(cat ${OS_PUBLIC_KEY2})
...
I1012 12:15:41.208693  292650 main.go:192] Found latest versions: OS 0.0.0+e4c55032fe1aadadaca7b752171966725c9a4d06, Applet 0.0.0+2e11bba36950d330bd2639463282f5448a66cc04
I1012 12:15:41.208715  292650 main.go:164] Fetching TRUSTED_OS bin from "trusted-os/0.0.0+e4c55032fe1aadadaca7b752171966725c9a4d06/trusted_os.elf"
I1012 12:15:41.211625  292650 main.go:203] Found OS bundle @ 34
I1012 12:15:41.211653  292650 main.go:164] Fetching TRUSTED_APPLET bin from "trusted-applet/0.0.0+2e11bba36950d330bd2639463282f5448a66cc04/trusted_applet.elf"
I1012 12:15:41.215448  292650 main.go:208] Found Applet bundle @ 19
I1012 12:15:41.215463  292650 main.go:164] Fetching BOOTLOADER bin from "boot/0.0.0+19548fd9325b95e81d4160ea7e38c0c1e3638f65/armored-witness-boot.imx"
I1012 12:15:41.216233  292650 main.go:213] Found Bootloader bundle @ 35
I1012 12:15:41.216242  292650 main.go:164] Fetching RECOVERY bin from "recovery/0.0.0+master/armory-ums.imx"
I1012 12:15:41.216797  292650 main.go:218] Found Recovery bundle @ 17
I1012 12:15:41.216803  292650 main.go:220] Loaded firmware artefacts.
```

The tool will prompt you to ensure that a device is plugged in and has the boot switch set correctly:

```bash
I1012 12:15:41.216806  292650 main.go:226] Operator, please ensure boot switch is set to USB (towards RJ45 socket), and then connect unprovisioned device 🙏
I1012 12:15:41.216809  292650 main.go:314] Waiting for device to be detected...
```

When it detects a device, it will continue with the flashing process:

```bash
I1012 12:15:42.276533  292650 sdp.go:85] found device 15a2:007d Freescale SemiConductor Inc  SE Blank 6UL
I1012 12:15:42.276573  292650 main.go:233] ✅ Detected device "/dev/hidraw0"
I1012 12:15:45.972100  292650 sdp.go:111] Attempting to SDP boot device /dev/hidraw0
I1012 12:15:45.972166  292650 sdp.go:123] Loading DCD at 0x00910000 (976 bytes)
I1012 12:15:45.975254  292650 sdp.go:128] Loading imx to 0x8000f400 (1793024 bytes)
I1012 12:15:46.951513  292650 sdp.go:133] Sending jump address to 0x8000f400
I1012 12:15:46.951744  292650 sdp.go:138] Serial download on /dev/hidraw0 complete
I1012 12:15:46.952068  292650 main.go:251] ✅ Witness device booting recovering image
I1012 12:15:46.952088  292650 main.go:379] Waiting for block device to appear
I1012 12:15:48.838362  292650 main.go:258] ✅ Detected blockdevice /dev/sdc
I1012 12:15:48.838403  292650 main.go:261]   Flashing in 5
I1012 12:15:49.838553  292650 main.go:261]   Flashing in 4
...
```

Once complete, it will ask you to flip the boot swith and reboot the device:

```bash
I1012 12:18:10.816923  292650 main.go:278] Operator, please change boot switch to MMC (away from RJ45 socket), and then reboot device 🙏
I1012 12:18:10.816955  292650 main.go:279] Waiting for device to boot...
I1012 12:18:10.816957  292650 main.go:337] Waiting for armored witness device to be detected...
```

When you've done this, the device should boot into the witness firmware which the `provision` tool will detect:

```bash
I1012 12:26:27.081822  292650 main.go:287] ✅ Detected device "/dev/hidraw0"
I1012 12:26:27.088214  292650 main.go:293] ✅ Witness serial number CA6B65D9D4992516 found
I1012 12:26:27.088249  292650 main.go:297] ✅ Witness serial number CA6B65D9D4992516 is not HAB fused
I1012 12:26:27.088274  292650 main.go:305] ✅ Witness ID DEV:ArmoredWitness-dawn-moon+271aa3a3+Abnd4ZwWVrpW9ioej/UDgP1YUaWI94YmIJPJHcXocnLM provisioned
I1012 12:26:27.088576  292650 main.go:100] ✅ Device provisioned!
```
