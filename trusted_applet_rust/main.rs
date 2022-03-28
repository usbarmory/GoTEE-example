// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// This is a minimal example for a GoTEE Trusted Applet written in Rust.
//
// For simplicity the example does not use any external crates or memory
// allocator, though in a complex applet both might be desirable/required.

#![no_std]
#![no_main]
#![feature(asm)]

use core::fmt::{self, Write};
use core::panic::PanicInfo;
use core::time::Duration;

const SYS_EXIT: u32 = 0;
const SYS_WRITE: u32 = 1;
const SYS_NANOTIME: u32 = 2;
const SYS_GETRANDOM: u32 = 3;

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
    log!("PL0 panic, {:?})", panic_info);
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
            "swi 0",
            in("r0") SYS_EXIT
        );
    }
}

impl Write for Stdout {
    fn write_str(&mut self, s: &str) -> fmt::Result {
        for c in s.bytes() {
            unsafe {
                asm!(
                    "swi 0",
                    in("r0") SYS_WRITE,
                    in("r1") c,
                );
            }
        }
        Ok(())
    }
}

fn nanotime() -> u64 {
    let ns_low: u32;
    let ns_high: u32;
    let ns: u64;

    unsafe {
        asm!(
            "swi 0",
            in("r0") SYS_NANOTIME,
        );

        asm!(
            "",
            out("r0") ns_low,
            out("r1") ns_high,
        );
    }

    ns = ((ns_high as u64) << 32) | (ns_low as u64);
    ns
}

fn getrandom(data: &mut [u8]) {
    unsafe {
        asm!(
            "swi 0",
            in("r0") SYS_GETRANDOM,
            in("r1") data.as_ptr(),
            in("r2") data.len(),
        );
    }
}

fn test_rng() {
    let mut rng: [u8; 16] = [0; 16];

    getrandom(&mut rng);

    log!(
        "PL0 obtained {:} random bytes from PL1: {:x?}",
        rng.len(),
        rng
    );
}

#[no_mangle]
pub extern "C" fn _start() {
    log!("PL0 rust/arm • TEE user applet (Secure World)");

    // test syscall interface
    test_rng();

    // terminate applet
    exit();
}
