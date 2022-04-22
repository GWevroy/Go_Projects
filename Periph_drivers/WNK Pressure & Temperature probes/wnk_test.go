// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wnk

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"periph.io/x/conn/v3/i2c/i2ctest"
	"periph.io/x/conn/v3/physic"
)

const (
	presRange       = 100 // Arbitrary pressure range selected for test (0 - 100kPa)
	presMin   int16 = -18 // Minimum pressure (boundary limit for error detection)
	presMax   int16 = 135 // Maximum pressure (boundary limit for error detection)
)

// Test #5 - Read Temperature.
func TestDev_ReadTemperature(t *testing.T) {
	// Variable function tests readout against test pressure.
	checkPres := func(testPres physic.Pressure, datH, datM, datL byte) {
		bus := i2ctest.Playback{
			Ops: []i2ctest.IO{
				{
					Addr: 0x6d,                     // WNK Device address.
					W:    []byte{0x06},             // Adress Pressure register).
					R:    []byte{datH, datM, datL}, // Read temperature.
				},
			},
		}

		defer bus.Close()
		bus.DontPanic = true

		d, err := NewSensorWNK(presRange, presMin, presMax, &bus, &DefaultOpts)
		if err != nil {
			t.Fatal(err)
		}

		testPres = testPres * physic.MilliPascal // Normalize test pressure

		Pressure, err := d.ReadPressure() // Call function under test

		strError := "" // Manage slices of error strings.
		if err != nil {
			if len(err.Error()) >= 22 {
				strError = err.Error()[:22]
			}

			// Detect pressure boundary has been exceeded and validly detected (mask this valid error).
			if strError != "pressure out of bounds" {
				t.Fatal(err) // Handle all non-boundary errors
			} else if (testPres > DefaultOpts.minPres) && (testPres < DefaultOpts.maxPres) {
				fmt.Printf("Pressure = %v  Min Press = %v  Max Press = %v\n", Pressure, DefaultOpts.minPres, DefaultOpts.maxPres)
				t.Fatal(errors.New("\nPressure incorrectly determined to be out of bounds"))
			}
		}

		// Detect if pressure boundary has been exceeded but not detected.
		if strError == "" && ((Pressure < DefaultOpts.minPres) || (Pressure > DefaultOpts.maxPres)) {
			t.Fatal(errors.New("\nPressure bounds trap failed to detect pressure out of range (" + Pressure.String() + ")"))
		}

		// Eliminate minor discrepancy (tolerance) for pressure comparisons
		if physic.Pressure(math.Abs(float64(Pressure-testPres))) < 10*physic.MilliPascal {
			Pressure = testPres
		}

		if Pressure.String() != testPres.String() && strError == "" {
			t.Fatalf("\nReadPressure() function test failed. Received %v, expected %v", Pressure, testPres)
		}
	}

	// Test range of temperatures (in milliCelcius)
	checkPres(136000000, 124, 229, 159) // 127°C (Test Upper bound).
	checkPres(133000000, 122, 145, 215) // 100kPa
	checkPres(100000000, 96, 248, 62)   // 100kPa
	checkPres(1500000, 20, 142, 190)    // 1.5kPa
	checkPres(0, 19, 100, 217)          // 0kPa
	checkPres(-2500000, 17, 116, 100)   // -2.5kPa
	checkPres(-10000000, 11, 163, 0)    // -10kPa
	checkPres(-20000000, 0, 0, 0)       // -20kPa  (Test Lower bound).
}

// Test #6 - String and Halt functions.
func TestDev_String(t *testing.T) {
	bus := i2ctest.Playback{}
	defer bus.Close()

	d, err := NewSensorWNK(presRange, presMin, presMax, &bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if s := d.String(); s != "WNK-Pressure" {
		t.Fatal(s)
	}
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
}
