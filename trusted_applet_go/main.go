// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"runtime"
	"runtime/goos"
	"time"

	"github.com/usbarmory/GoTEE/applet"
	"github.com/usbarmory/GoTEE/syscall"

	"github.com/usbarmory/GoTEE-example/mem"
	"github.com/usbarmory/GoTEE-example/util"
)

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)

	// yield to monitor (w/ err != nil) on runtime panic
	goos.Exit = applet.Crash
}

func testRNG(n int) {
	buf := make([]byte, n)
	syscall.GetRandom(buf, uint(n))
	log.Printf("applet obtained %d random bytes from monitor: %x", n, buf)
}

func testRPC() {
	res := ""
	req := "hello"

	log.Printf("applet requests echo via RPC: %s", req)
	err := syscall.Call("RPC.Echo", req, &res)

	if err != nil {
		log.Printf("applet received RPC error: %v", err)
	} else {
		log.Printf("applet received echo via RPC: %s", res)
	}
}

func main() {
	log.Printf("%s/%s (%s) • TEE user applet", runtime.GOOS, runtime.GOARCH, runtime.Version())

	// test syscall interface
	testRNG(16)

	// test RPC interface
	testRPC()

	log.Printf("applet will sleep for 5 seconds")

	ledStatus := util.LEDStatus{
		Name: "blue",
		On:   true,
	}

	// test concurrent execution of applet and supervisor/monitor
	for i := 0; i < 5; i++ {
		syscall.Call("RPC.LED", ledStatus, nil)
		ledStatus.On = !ledStatus.On

		time.Sleep(1 * time.Second)
		log.Printf("applet says %d mississippi", i+1)
	}

	// test memory protection
	mem.TestAccess("applet")

	// this should be unreachable

	// test exception handling
	mem.TestDataAbort("applet")

	// terminate applet
	applet.Exit()
}
