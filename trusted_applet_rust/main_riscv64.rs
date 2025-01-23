// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// This is a minimal example for a GoTEE Trusted Applet written in Rust.
//
// For simplicity the example does not use any external crates or memory
// allocator, though in a complex applet both might be desirable/required.

#![no_std]
#![no_main]

use core::arch::asm;
use core::fmt::{self, Write};
use core::panic::PanicInfo;
use core::time::Duration;

const SYS_EXIT: u64 = 0;
const SYS_WRITE: u64 = 1;
const SYS_NANOTIME: u64 = 2;
const SYS_GETRANDOM: u64 = 3;

struct Stdout {}

macro_rules! print {
    ($($arg:tt)*) => {
        write!(&mut Stdout {}, $($arg)*).ok();
    };
}

macro_rules! log {
    ($fmt:expr) => {
        log_walltime(nanotime());
        print!(concat!($fmt, "\r\n"))
    };
    ($fmt:expr, $($arg:tt)*) => {
        log_walltime(nanotime());
        print!(concat!($fmt, "\r\n"), $($arg)*)
    };
}

#[panic_handler]
fn panic(panic_info: &PanicInfo) -> ! {
    log!("applet panic, {:?})", panic_info);
    exit();

    // this should be unreachable
    loop {}
}

fn log_walltime(ns: u64) {
    let epoch = Duration::from_nanos(ns).as_secs();
    let ss = epoch % 60;
    let mm = (epoch / 60) % 60;
    let hh = (epoch / 3600) % 24;

    print!("{:02}:{:02}:{:02} ", hh, mm, ss);
}

fn exit() {
    unsafe {
        asm!(
            "ecall",
            in("a0") SYS_EXIT,
        );
    }
}

impl Write for Stdout {
    fn write_str(&mut self, s: &str) -> fmt::Result {
        for c in s.bytes() {
            unsafe {
                asm!(
                    "ecall",
                    in("a0") SYS_WRITE,
                    in("a1") c,
                    in("a7") 0,
                );
            }
        }
        Ok(())
    }
}

fn nanotime() -> u64 {
    let ns: u64;

    unsafe {
        asm!(
            "ecall",
            in("a0") SYS_NANOTIME,
            in("a7") 0,
        );

        asm!(
            "",
            out("a0") ns,
        );
    }

    ns
}

fn getrandom(data: &mut [u8]) {
    unsafe {
        asm!(
            "ecall",
            in("a0") SYS_GETRANDOM,
            in("a1") data.as_ptr(),
            in("a2") data.len(),
            in("a7") 0,
        );
    }
}

fn test_rng() {
    let mut rng: [u8; 16] = [0; 16];

    getrandom(&mut rng);

    log!(
        "applet obtained {:} random bytes from SM: {:x?}",
        rng.len(),
        rng
    );
}

#[no_mangle]
pub extern "C" fn _start() {
    log!("rust • TEE user applet");

    // test syscall interface
    test_rng();

    // terminate applet
    exit();
}
