# ArmoredWitness

The ArmoredWitness project is intended to kick-start a cross-ecosystem witness network, providing split-view attack prevention to a growing set of transparency-enabled ecosystems.

## Background

Transparency systems work by ensuring that all actors in a given ecosystem see the same append-only list of data, typically stored in [verifiable logs](https://transparency.dev/verifiable-data-structures/). This allows folks relying on this data to be confident that even if they are unable to determine the correctness of the data themselves, it is visible to others who _can_ verify it and call out badness.

If a malicious log is able to present inconsistent views of itself to different roles, then such badness will go undetected.

Witnessing is a solution to this "split-view" attack - well-known identities verify the append-only property of the logs; countersigning checkpoints (commitments to the state of a log) which have been proved to be consistent with all earlier checkpoints from the same log which have also been seen by the witness.

## Goals

We want to:

* **Help transparency-enabled ecosystems further tighten their security properties.** \
By providing a lightweight network which is compatible with ecosystems using a [common checkpoint format](https://github.com/transparency-dev/formats/tree/main/log), we can help reduce the trust being placed in log operators.  \
Existing compatible ecosystems include: Go's sum DB, Sigstore, Pixel Binary Transparency, LVFS, SigSum, ArmoryDrive.
* **Ensure the ArmoredWitness device is as low-touch and maintenance-free as possible.** \
We don't want to ask anyone to be a full-time system administrator for these devices, they should be as "plug-in and go" as possible.
* **Bootstrap a diverse witnessing ecosystem** \
Encourage others to participate, learn with us, and potentially even go on to develop their own witness protocols and networks, ideally in such a way that interoperability remains possible and enables greater diversification of reputational trust.

## Device

The ArmoredWitness device is a small networked device based on the [USBArmory](https://inversepath.com/usbarmory) that runs an open-source implementation of a witness configured to support a growing number of transparency-enabled ecosystems.

![alt_text](images/armored-witness-render.png "ArmoredWitness case render")

We're making a small number of these devices, and plan to distribute them to folks who are passionate about one or more of the witnessed ecosystems.

The ArmoredWitness devices will only run the ArmoredWitness firmware, so will not be repurposable for other use cases.

TODO: We'll add more information about the hardware

## Software

The firmware for the ArmoredWitness is all written in Go, and compiled with [TamaGo](https://github.com/usbarmory/tamago) into a bare-metal executable. This enables us to take advantage of Go's memory safety and excellent standard library, and avoid needing to take a dependency on any traditional generic bootloader/kernel/OS combinations, considerably reducing the surface area of the codebase.

There are 3 main parts to the ArmoredWitness firmware stack:

* Bootloader \
<https://github.com/transparency-dev/armored-witness-bootloader> \
A very simple TamaGo unikernel which loads the OS from MMC, verifies it, and finally boots it.
* Trusted OS \
<https://github.com/transparency-dev/armored-witness-os> \
The OS is a TamaGo unikernel which is primarily concerned with:
  * Managing the device hardware (Ethernet, storage, LEDs, etc.),
  * Loading and verifying the Applet from MMC, and executing it inside a TEE,
  * Providing an RPC-like syscall interface to the Applet.
* Witness Applet \
<https://github.com/transparency-dev/armored-witness-applet>  \
The Trusted Applet is where the witness itself lives. \
In addition to running the witness code, the applet also handles the TCP/IP side of networking (using RPC to send/receive packets).

Along with other ancillary parts:

* Tooling \
This repo contains some tooling to help provision the factory-fresh devices into ArmoredWitness devices.
* Recovery Image \
<https://github.com/usbarmory/armory-ums>
