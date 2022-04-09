// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package max17040

import (
	"testing"

	"periph.io/x/conn/v3/i2c/i2ctest"
	"periph.io/x/conn/v3/physic"
)

// Test #1 - String and Halt functions.
func TestDev_String(t *testing.T) {
	bus := i2ctest.Playback{}
	defer bus.Close()

	d, err := NewMAX17040(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if s := d.String(); s != "MAX17040" {
		t.Fatal(s)
	}
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
}

// Test #2 - Device Reset.
func TestDev_Reset(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: 0x36,                     // MAX1704x Device address.
				W:    []byte{0xfe, 0x00, 0x54}, //Reference (Command) Register 0xfe. Write 0x0054 to initiate POR.
				R:    []byte{},
			},
		},
	}
	defer bus.Close()
	bus.DontPanic = true

	d, err := NewMAX17040(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}

	err = d.Reset()
	if err != nil {
		t.Fatal(err)
	}

}

// Test #3 - SoC Algorithm QuickStart.
func TestDev_QuickStart(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: 0x36,                     // MAX1704x Device address.
				W:    []byte{0x06, 0x40, 0x00}, //Reference (Mode) Register 0x06. Write 0x4000 to initiate Quick Start.
				R:    []byte{},
			},
		},
	}
	defer bus.Close()
	bus.DontPanic = true

	d, err := NewMAX17040(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}

	err = d.QuickStart()
	if err != nil {
		t.Fatal(err)
	}

}

// Test #4 - Set RCOMP calibration value.
func TestDev_SetRCOMP(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: 0x36,                     // MAX1704x Device address.
				W:    []byte{0x0c, 0x34, 0x56}, //Reference (RCOMP) Register 0x0c. Write arbitrary value.
				R:    []byte{},
			},
		},
	}
	defer bus.Close()
	bus.DontPanic = true

	const targetVal = 0x3456 // Result expected to be returned via I2C.

	d, err := NewMAX17040(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}

	err = d.SetRCOMP(targetVal)
	if err != nil {
		t.Fatal(err)
	}
}

// Test #5 - Get RCOMP calibration value.
func TestDev_GetRCOMP(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: 0x36,               // MAX1704x Device address.
				W:    []byte{0x0c},       //Reference (RCOMP) Register 0x0c.
				R:    []byte{0x97, 0x00}, // Read arbitrary value (must fall within boundary limits though).
			},
		},
	}
	defer bus.Close()
	bus.DontPanic = true

	const expectedVal = 0x9700 // Result expected to be returned via I2C.

	d, err := NewMAX17040(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}

	valRComp, err := d.GetRCOMP()
	if err != nil {
		t.Fatal(err)
	}
	if valRComp != expectedVal {
		t.Fatalf("GetRCOMP() function test received %d, expected %d", valRComp, expectedVal)
	}
}

// Test #6 - Fetch Device Version.
func TestDev_GetVersion(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: 0x36,               // MAX1704x Device address.
				W:    []byte{0x08},       // Reference (Version) Register.
				R:    []byte{0x40, 0x50}, // Read arbitrary value (Must tally with expectedVer).
			},
		},
	}
	defer bus.Close()
	bus.DontPanic = true

	const expectedVer = "64/80" // Result expected to be returned via I2C.

	d, err := NewMAX17040(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}

	Version, err := d.GetVersion()
	if err != nil {
		t.Fatal(err)
	}
	if Version != expectedVer {
		t.Fatalf("GetVersion() function test received %v, expected %v", Version, expectedVer)
	}
}

// Test #7 - Read Cell DC Voltage.
func TestDev_ReadCellVoltage(t *testing.T) {
	var targetV = []byte{0x98, 0x10} // Result expected to be returned via I2C (Least significant nibble is don't care).

	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: 0x36,                           // MAX1704x Device address.
				W:    []byte{0x02},                   //Reference (Voltage) Register 0x02.
				R:    []byte{targetV[0], targetV[1]}, // Read arbitrary value (must fall within boundary limits).
			},
		},
	}
	defer bus.Close()
	bus.DontPanic = true

	d, err := NewMAX17040(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}

	Volts, err := d.ReadCellVoltage()
	if err != nil {
		t.Fatal(err)
	}

	exptectedV := physic.ElectricPotential((int(targetV[0])<<4 | int(targetV[1])>>4)) * 1250000 // Calc expected volts.

	if Volts != exptectedV {
		t.Fatalf("ReadCellVoltage() function test received %v, expected %v", Volts, exptectedV)
	}
}

// Test #8 - Read State of Charge.
func TestDev_ReadSOC(t *testing.T) {
	var targetSOC = []byte{0x60, 0x83} // Result expected to be returned via I2C.

	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: 0x36,                               // MAX1704x Device address.
				W:    []byte{0x04},                       // Reference (State Of Charge) Register 0x04.
				R:    []byte{targetSOC[0], targetSOC[1]}, // Read arbitrary value (must fall within 0 - 100% boundary).
			},
		},
	}
	defer bus.Close()
	bus.DontPanic = true

	d, err := NewMAX17040(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}

	SoC, err := d.ReadSoC()
	if err != nil {
		t.Fatal(err)
	}

	expectedSOC := float32(uint16(targetSOC[0])<<8|uint16(targetSOC[1])) / 256 // (Calculate expected State of Charge %).

	if SoC != expectedSOC {
		t.Fatalf("ReadSoC() function test received %v%%, expected %v%%", SoC, expectedSOC)
	}
}
