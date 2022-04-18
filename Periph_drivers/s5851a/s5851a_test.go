// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package s5851a

import (
	"errors"
	"strings"
	"testing"

	"periph.io/x/conn/v3/i2c/i2ctest"
	"periph.io/x/conn/v3/physic"
)

const (
	ptrTemp byte = 0
	ptrConf byte = 1
)

// Test #1 - Enter Sleep (Shutdown) Mode.
func TestDev_Shutdown(t *testing.T) {
	// Prep Mock I2C
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: I2CAddr3,           // S-5851A Device address.
				W:    []byte{ptrConf, 1}, // Expect (write) pointer register points to Config register, set shutdown bit)
				R:    []byte{},
			},
		},
	}

	defer bus.Close()
	bus.DontPanic = true

	d, err := NewS5851A(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	err = d.Shutdown(true) // Initiate shutdown.
	if err != nil {
		t.Fatal(errors.New("\nshutdown failed. " + err.Error()))
	}
}

// Test #2 - Wake from Sleep (Shutdown) Mode.
func TestDev_WakeFromSleep(t *testing.T) {
	//Prep Mock I2C
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: I2CAddr3,           // S-5851A Device address.
				W:    []byte{ptrConf, 0}, // Expect (write) pointer register points to Config register, clear shutdown bit)
				R:    []byte{},
			},
		},
	}

	defer bus.Close()
	bus.DontPanic = true

	d, err := NewS5851A(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	err = d.Shutdown(false) // Wake up.
	if err != nil {
		t.Fatal(errors.New("\nwake from sleep failed. " + err.Error()))
	}
}

// Test #3 - Initiate a single temperature reading.
func TestDev_OneShotTrigger(t *testing.T) {
	//
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: I2CAddr3,             // S-5851A Device address.
				W:    []byte{ptrConf, 129}, // Expect (write) pointer register points to Config register, set trigger bits).
				R:    []byte{},
			},
		},
	}

	defer bus.Close()
	bus.DontPanic = true

	d, err := NewS5851A(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	err = d.OneShotTrigger() // Trigger a single temperature reading, then sleep.
	if err != nil {
		t.Fatal(errors.New("\ntrigger of one shot temperature read failed. " + err.Error()))
	}
}

// Test #4 - poll for completed temperature data acquisition.
func TestDev_IsTempDone(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: I2CAddr3,        // S-5851A Device address.
				W:    []byte{ptrConf}, // Point to config register.
				R:    []byte{},
			},
			{
				Addr: I2CAddr3,
				W:    []byte{},
				R:    []byte{128}, // Read config status (Bit 7 = 1 = still busy).
			},
			{
				Addr: I2CAddr3,        // S-5851A Device address.
				W:    []byte{ptrConf}, // Point to config register.
				R:    []byte{},
			},
			{
				Addr: I2CAddr3,
				W:    []byte{},
				R:    []byte{127}, // Read config status (Bit 7 = 0 = temperature ready to read).
			},
		},
	}
	defer bus.Close()
	bus.DontPanic = true

	// Test detection of signal (temperature acquisition still busy).
	d, err := NewS5851A(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	isFinished, err := d.IsTempReady()
	if err != nil {
		t.Fatal(errors.New("\nfailed to determine config status. " + err.Error()))
	}

	if isFinished {
		t.Fatal(errors.New("Failed to receive (temperature acquisition busy) signal"))
	}

	// Test detection of signal (conversion complete, ready to read).
	isFinished, err = d.IsTempReady()
	if err != nil {
		t.Fatal(errors.New("\nfailed to determine config status. " + err.Error()))
	}

	if !isFinished {
		t.Fatal(errors.New("Failed to receive (temperature acquisition complete, ready to read) signal"))
	}
}

// Test #5 - Read Temperature.
func TestDev_ReadTemperature(t *testing.T) {

	// On startup driver should attempt to reset pointer register.
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: I2CAddr3,  // S-5851A Device address.
				W:    []byte{0}, // Write pointer reference (back to temperature read register).
				R:    []byte{},
			},
			{
				Addr: I2CAddr3,     // S-5851A Device address.
				W:    []byte{},     // Empty write (S-5851A pointer already points to temperature register).
				R:    []byte{0, 0}, // Dummy (Don't care) temperature for this test.
			},
		},
	}
	bus.DontPanic = true

	d, err := NewS5851A(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.ReadTemperature()

	if err != nil {
		t.Fatal(errors.New("\npointer register not initialized. " + err.Error()))
	}

	bus.Close() // close bus prior to new bus being created for temperature readouts about to follow.

	// Variable function tests readout against test temperature.
	// Temperature data-points taken from S-5851 datasheet for ease of use.
	checkTemp := func(testTemp physic.Temperature, datH, datL byte) {
		bus := i2ctest.Playback{
			Ops: []i2ctest.IO{
				{
					Addr: I2CAddr3,           // S-5851A Device address.
					W:    []byte{},           // Empty write (S-5851A pointer already points to temperature register).
					R:    []byte{datH, datL}, // Read temperature.
				},
			},
		}

		defer bus.Close()
		bus.DontPanic = true

		d, err := NewS5851A(&bus, &DefaultOpts)
		if err != nil {
			t.Fatal(err)
		}

		Temperature, err := d.ReadTemperature()

		strError := "" // Manage slices of error strings.
		if err != nil {
			if len(err.Error()) >= 25 {
				strError = strings.Clone(err.Error()[:25]) // Make physical copy of error substring (not reference!).
			}

			// Detect temperature boundary has been exceeded and validly detected (mask this valid error).
			if strError != "temperature out of bounds" {
				t.Fatal(err) // Handle all non-boundary errors
			} else if Temperature > minimumTemp && Temperature < maximumTemp {
				t.Fatal(errors.New("\nTemperature incorrectly determined to be out of bounds"))
			}
		}
		// Detect if temperature boundary has been exceeded but not detected.
		if strError != "temperature out of bounds" && ((Temperature < minimumTemp) || (Temperature > maximumTemp)) {
			t.Fatal(errors.New("\nTemperature bounds trap failed to detect temperature out of range (" + Temperature.String() + ")"))
		}

		exptectedT := physic.Temperature(testTemp*physic.MilliKelvin + 273150000000)

		// Verify measured temperature is the expected temperature (Boundary violations are masked to eliminate false reads)
		if Temperature != exptectedT && strError != "temperature out of bounds" {
			t.Fatalf("\nReadTemperature() function test failed. Received %v, expected %v", Temperature, exptectedT)
		}
	}

	// Test range of temperatures (in milliCelcius)
	checkTemp(127000, 0x7f, 0x00) // 127°C (Test Upper bound).
	checkTemp(100000, 0x64, 0x00) // 100°C
	checkTemp(75000, 0x4b, 0x00)  // 75°C
	checkTemp(25000, 0x19, 0x00)  // 25°C
	checkTemp(250, 0x00, 0x40)    // 0.25°C
	checkTemp(0, 0x00, 0x00)      // 0°C
	checkTemp(-250, 0xff, 0xc0)   // -0.25°C
	checkTemp(-25000, 0xe7, 0x00) // -25°C
	checkTemp(-35000, 0xdd, 0x00) // -35°C
	checkTemp(-47000, 0xd1, 0x00) // -47°C (Test Lower bound).
}

// Test #6 - String and Halt functions.
func TestDev_String(t *testing.T) {
	bus := i2ctest.Playback{}
	defer bus.Close()

	d, err := NewS5851A(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if s := d.String(); s != "S-5851A" {
		t.Fatal(s)
	}
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
}
