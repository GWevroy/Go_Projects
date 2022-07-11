// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wnk

import (
	"errors"
	"math"
	"testing"

	"periph.io/x/conn/v3/i2c/i2ctest"
	"periph.io/x/conn/v3/physic"
)

const (
	presRange                    = 100                           // Arbitrary pressure range selected for test (0 - 100kPa)
	presMin   int16              = -18                           // Minimum pressure (boundary limit for error detection)
	presMax   int16              = 135                           // Maximum pressure (boundary limit for error detection)
	ZeroDegC  physic.Temperature = -273150 * physic.MilliKelvin  // Zero Degrees Celcius expressed in Degrees milliKelvin (offset)
	tempMin   physic.Temperature = -40*physic.Celsius - ZeroDegC // Default Minimum temperature
	tempMax   physic.Temperature = 125*physic.Celsius - ZeroDegC // Default Maximum temperature

	TestI2Caddr  uint16 = 0x6d // I2C Address of device under test
	presRegister byte   = 0x06 // Transducer Pressure (read) register
	TempRegister byte   = 0x09 // Transducer Temperature (read) register
)

// Test #1 - Read Temperature.
func TestDev_Temperature(t *testing.T) {
	// Variable function tests readout against test temperature.
	checkTemp := func(testTemp physic.Temperature, datH, datM, datL byte) {
		bus := i2ctest.Playback{
			Ops: []i2ctest.IO{
				{
					Addr: TestI2Caddr,              // WNK Device address.
					W:    []byte{TempRegister},     // Register: Temperature.
					R:    []byte{datH, datM, datL}, // Read temperature.
				},
			},
		}

		defer bus.Close()
		bus.DontPanic = true

		d, err := NewSensorWNK(0, 0, 0, &bus, &DefaultOpts)
		if err != nil {
			t.Fatal(err)
		}

		testTemp = testTemp*physic.MilliCelsius - ZeroDegC // Normalize test temperature

		Temperature, err := d.ReadTemperature() // Call function under test

		strError := "" // Prepare to filter out (out of bounds) error
		if err != nil {
			if len(err.Error()) > 25 {
				strError = err.Error()[:25]
			}

			// Detect temperature boundary has been exceeded and validly detected (mask this valid error).
			if strError != "temperature out of bounds" {
				t.Fatal(err) // Handle all non-boundary errors
			} else if (testTemp > tempMin) && (testTemp < tempMax) {
				t.Fatal(errors.New("\nTemperature incorrectly determined to be out of bounds (" + testTemp.String() + ")"))
			}
		}

		// Detect if pressure boundary has been exceeded but not detected.
		if strError == "" && ((Temperature < tempMin) || (Temperature > tempMax)) {
			t.Fatal(errors.New("\nTemperature bounds trap failed to detect temperature out of range (" + Temperature.String() + ")"))
		}

		// Eliminate minor discrepancy (tolerance) for temperature comparisons
		if physic.Temperature(math.Abs(float64(Temperature-testTemp))) < 10*physic.MilliCelsius {
			Temperature = testTemp
		}

		if Temperature.String() != testTemp.String() && strError == "" {
			t.Fatalf("\nReadTemperature() function test failed. Received %v, expected %v", Temperature, testTemp)
		}
	}

	// Test range of temperatures (m°C)
	checkTemp(130000, 105, 0, 0)    // 130°C (Test Upper bound).
	checkTemp(25000, 0, 0, 0)       // 25°C
	checkTemp(5300, 236, 76, 204)   // 5.3°C
	checkTemp(0, 231, 0, 0)         // 0°C
	checkTemp(-390, 230, 156, 40)   // -0.39°C
	checkTemp(-17850, 213, 38, 102) // -17.85°C
	checkTemp(-27530, 203, 120, 81) // -27.53°C (Test Lower bound).
}

// Test #2 - Read Pressure.
func TestDev_Pressure(t *testing.T) {
	// Variable function tests readout against test pressure.
	checkPres := func(testPres physic.Pressure, datH, datM, datL byte) {
		bus := i2ctest.Playback{
			Ops: []i2ctest.IO{
				{
					Addr: TestI2Caddr,              // WNK Device address.
					W:    []byte{0x06},             // Register: Pressure.
					R:    []byte{datH, datM, datL}, // Read pressure.
				},
			},
		}

		defer bus.Close()
		bus.DontPanic = true

		d, err := NewSensorWNK(presRange, presMin, presMax, &bus, &DefaultOpts)
		if err != nil {
			t.Fatal(err)
		}

		testPres = testPres * physic.Pascal // Normalize test pressure

		Pressure, err := d.ReadPressure() // Call function under test

		strError := "" // Prepare to filter out single error.
		if err != nil {
			if len(err.Error()) > 22 {
				strError = err.Error()[:22]
			}

			// Detect pressure boundary has been exceeded and validly detected (mask this valid error).

			if strError != "pressure out of bounds" {
				t.Fatal(err) // Handle all non-boundary errors
			} else if (testPres > DefaultOpts.minPres) && (testPres < DefaultOpts.maxPres) {
				t.Fatal(errors.New("\nPressure incorrectly determined to be out of bounds" + testPres.String() + ")"))
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

	// Test range of pressures (Pascals)
	checkPres(136000, 124, 229, 159) // 136kPa (Test Upper bound).
	checkPres(133000, 122, 145, 215) // 100kPa
	checkPres(100000, 96, 248, 62)   // 100kPa
	checkPres(1500, 20, 142, 190)    // 1.5kPa
	checkPres(0, 19, 100, 217)       // 0kPa
	checkPres(-2500, 17, 116, 92)    // -2.5kPa
	checkPres(-10000, 11, 162, 231)  // -10kPa
	checkPres(-20000, 3, 224, 247)   // -20kPa  (Test Lower bound).
}

// Test #3 - String and Halt functions.
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
