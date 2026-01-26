# This script must be invoked with the riscv64 compiler as first argument (e.g. build.sh riscv64-linux-gnu-gcc)
$1 -march=rv64g -mabi=lp64 -static -mcmodel=medany -fvisibility=hidden -nostdlib -nostartfiles -Tbios.ld bios.s -o bios.bin
