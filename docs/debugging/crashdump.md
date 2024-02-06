# Debugging Crashbumps in the night

If something goes _very_ wrong in the applet, you may see a register dump like below:

```
00:00:21    r0:91a19b98  r1:91a163a0  r2:00000000  r3:00000000
00:00:21    r4:00000040  r5:91aea420  r6:91a6cfc0  r7:00000044
00:00:21    r8:c9ada987  r9:ada987ef r10:91803600 r11:00000594 cpsr:200001d7 (ABT)
00:00:21   r12:eec9ada9  sp:91a195cc  lr:901ec840  pc:901eca54 spsr:200001d0 (USR)
00:00:21 SM applet stopped mode:USR sp:0x91a195cc lr:0x901ec840 pc:0x901eca54 ns:false err:ABT
00:00:21
```

Using the value from the `pc` register (`0x901eca54` above), we can at least figure out where
the applet was executing when it crashed.

For this, you need to have the applet ELF file which was installed on the device at the time,
if it's a local development build this is easy, for CI or prod, it's _probably_ the most recent
entry in the corresponding Firmware Transparency log. Failing this, it should be possible to
modify the [`cmd/verify`](/cmd/verify) tool in this repo to either identify exactly which log
entry it was, or better still, dump the currently installed version to a local file.
(TODO(al): do this.)

With the ELF file at hand, we can use `Go`'s built-in tooling to disassemble the applet and
produce a file which will allow us to pin point exactly which line of code was executing at
the time the crash occurred.

First, run `go tool objdump ./path/to/applet.elf > /tmp/applet.od` this will produce a file
which looks something like this:

```
...
  net.go:365            0x2068750c              e1510000                CMP R0, R1
  net.go:365            0x20687510              dafffffb                B.LE 0x20687504
  net.go:365            0x20687514              ebf5f0ea                BL gvisor.dev/gvisor/pkg/tcpip/stack.New(SB)
  net.go:365            0x20687518              e59d0344                MOVW 0x344(R13), R0
  net.go:364            0x2068751c              e3a01000                MOVW $0, R1
  net.go:364            0x20687520              e59d235c                MOVW 0x35c(R13), R2
  net.go:364            0x20687524              e5821000                MOVW R1, (R2)
  net.go:364            0x20687528              e59fb2e0                MOVW 0x2e0(R15), R11
  net.go:364            0x2068752c              e59b1000                MOVW (R11), R1
  net.go:364            0x20687530              e3510000                CMP $0, R1
  net.go:364            0x20687534              0a000007                B.EQ 0x20687558
  net.go:364            0x20687538              e5921004                MOVW 0x4(R2), R1
  net.go:364            0x2068753c              ebe7e5cc                BL runtime.gcWriteBarrier4(SB)
  net.go:364            0x20687540              e5881000                MOVW R1, (R8)
...
```

Searching for the `pc` address from the crash _should_ end up matching the entry in the 2nd column above.
This will tell you exactly which machine instruction faulted (on the right), and which line of Go code
that instruction came from (first column, on the left).

## Addressing

Depending on which firmware crashed, and compilation options (specifically, `BEE=0` vs `BEE=1`), the address
ranges in use will be different:

| Firmware     | Memory start (BEE) | Memory start (non-BEE)
|--------------|--------------------|------------------------
| `Bootloader` |  N/A               | `0x90000000`
| `OS`         | `0x20000000`       | `0x90000000`
| `Applet`     | `0x20000000`       | `0x90000000`


