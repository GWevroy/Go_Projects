// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This driver has been hardware tested (pressure only, no temperature feature)
// on a WNK805 100kPa pressure sensor.
// Should be equally suitable to drive WNK21, WNK19, WNK811, WNK80mA, WNK8010
// transducers using I2C communications.

package wnk

import (
	"errors"
	"sync"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/physic"
)

// Device driver constants.
const (
	I2CAddr uint16 = 0x6d // I2CAddr1 is the default I2C device address for the WNK pressure/temperature transducer series.

	//maximumPres physic.Pressure = 120 * physic.KiloPascal // Maximum anticipated pressure (Upper error detection bound).
	//minimumPres physic.Pressure = -35 * physic.KiloPascal // Minimum anticipated pressure (Lower error detection bound).

	regPressure byte               = 0x06 // Transducer Pressure (read) register
	regTemp     byte               = 0x09 // Transducer Temperature (read) register
	maxTemp     physic.Temperature = 125  // Default Maximum temperature
	minTemp     physic.Temperature = -40  // Default Minimum temperature
)

var (
	data = []byte{0, 0, 0} // data is a container for I2C received data frames.
)

// Opts holds the configuration options.
type Opts struct {
	I2cAddress uint16 // I2C Address of device.
	maxPres    physic.Pressure
	minPres    physic.Pressure
	MinTemp    physic.Temperature // Absolute maximum temperature (upper boundary)
	MaxTemp    physic.Temperature // Absolute minimum temperature (lower boundary)
}

// DefaultOpts are the recommended default options.
var DefaultOpts = Opts{
	I2cAddress: I2CAddr,
	MinTemp:    minTemp,
	MaxTemp:    maxTemp,
}

// Dev is a handle to the device methods.
type Dev struct {
	c      i2c.Dev
	name   string
	pRange uint16     // Pressure range (in kPa). EG 0kPa to 500kPa = 500 - 0 = 500.
	mu     sync.Mutex // mu inhibits opportunities for simultaneous I2C operations.
}

// NewSensorWNK creates a new driver for the pressure/temperature transducer.
func NewSensorWNK(kPaRange int16, kPaMin int16, kPaMax int16, busI2C i2c.Bus, opts *Opts) (*Dev, error) {
	if kPaMax < (kPaMin+kPaRange) || (kPaMin > (kPaMax - kPaRange)) {
		return nil, errors.New("handle construct error for WNK-Pressure sensor (Pressure range out of bounds)")
	}
	opts.maxPres = physic.Pressure(kPaMax) * physic.KiloPascal // Set upper bound above which pressure a fatal error is triggered
	opts.minPres = physic.Pressure(kPaMin) * physic.KiloPascal // Set lower bound below which pressure a fatal error is triggered

	return &Dev{
		c:      i2c.Dev{Bus: busI2C, Addr: opts.I2cAddress},
		name:   "WNK-Pressure",
		pRange: uint16(kPaRange),
	}, nil
}

// ReadPressure reads pressure from WNK device.
func (d *Dev) ReadPressure() (pressure physic.Pressure, err error) {
	// Lock device to inhibit attempts at multiple simultaneous read/writes.
	d.mu.Lock()
	defer d.mu.Unlock()

	var presWorkings float64

	//Read raw (pressure) data from device.
	if err := d.c.Tx([]byte{regPressure}, data); err != nil {
		return 0, errors.New("failed to read pressure from " + d.name + ". " + err.Error())
	}

	tmpkPa := uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2]) // Big Endian (left shift MSB) integer created from bytes.
	if (tmpkPa & 0x800000) != 0 {
		presWorkings = float64(tmpkPa) - 16777216 // Adjust for negative (sensed) pressure if applicable.
	} else {
		presWorkings = float64(tmpkPa)
	}

	presWorkings = 3.3 * presWorkings / 8388608

	presWorkings = float64(d.pRange) * (presWorkings - 0.5) / 2

	pressure = physic.Pressure(presWorkings * 1000000000000) // Cast final pressure.

	// Provide some boundary (sanity) checks on pressure read.
	// Reads outside of range are considered unreliable and may indicate
	// either noisy comms or defective transducer.
	if (pressure > DefaultOpts.maxPres) || (pressure < DefaultOpts.minPres) {

		err = errors.New("pressure out of bounds. " + d.name + " transducer measured " + pressure.String())
		return 0, err
	}
	return pressure, nil
}

// ReadTemperature reads temperature from WNK device.
func (d *Dev) ReadTemperature() (temperature physic.Temperature, err error) {
	// Lock device to inhibit attempts at multiple simultaneous read/writes.
	d.mu.Lock()
	defer d.mu.Unlock()

	var tempWorkings float64

	//Read raw (temperature) data from device.
	if err := d.c.Tx([]byte{regTemp}, data); err != nil {
		return 0, errors.New("failed to read temperature from " + d.name + ". " + err.Error())
	}

	tmpkPa := uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2]) // Big Endian (left shift MSB) integer created from bytes.
	if (tmpkPa & 0x800000) != 0 {
		tempWorkings = float64(tmpkPa) - 16777216 // Adjust for negative (sensed) temperature if applicable.
	} else {
		tempWorkings = float64(tmpkPa)
	}

	tempWorkings = 25 + tempWorkings/65536

	temperature = physic.Temperature(tempWorkings * 1000000000000) // Cast final temperature.

	// Provide some boundary (sanity) checks on temperature read.
	// Reads outside of range are considered unreliable and may indicate
	// either noisy comms or defective transducer.
	if (temperature > DefaultOpts.MaxTemp) || (temperature < DefaultOpts.MinTemp) {

		err = errors.New("temperature out of bounds. " + d.name + " transducer measured " + temperature.String())
		return 0, err
	}
	return temperature, nil
}

// String implements conn.Resource.
func (d *Dev) String() string {
	return d.name
}

// Halt implements conn.Resource.
func (d *Dev) Halt() error {
	return nil
}
