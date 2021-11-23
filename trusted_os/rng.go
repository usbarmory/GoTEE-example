// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/canonical/go-sp800.90a-drbg"

	"github.com/f-secure-foundry/tamago/soc/imx6"
	"github.com/f-secure-foundry/tamago/soc/imx6/rngb"
)

// yieldRNGB re-configures the TamaGo runtime entropy source to a pure software
// one (NIST SP 800-90A DRBG) to allow scenarios where it is desirable to give
// exclusive RNGB access to the Normal World OS.
func yieldRNGB() {
	seed := make([]byte, 256)
	rngb.GetRandomData(seed)

	nonce := make([]byte, 128)
	rngb.GetRandomData(nonce)

	uid := imx6.UniqueID()

	rng, err := drbg.NewCTRWithExternalEntropy(32, seed, nonce, uid[:], nil)

	if err != nil {
		panic(fmt.Sprintf("could not instantiate DRBG, %v", err))
	}

	// override TamaGo entropy source with an RNGB seeded DRGB
	imx6.SetRNG(func(b []byte) {
		rng.Read(b)
	})

	rngb.Reset()
}

// restoreRNGB restores the RNGB as TamaGo runtime entropy source.
func restoreRNGB() {
	imx6.SetRNG(rngb.GetRandomData)
}
