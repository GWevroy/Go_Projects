// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package s5851a

import (
	"errors"
	"sync"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/physic"
)

// Device driver constants.
const (
	I2CAddr1 uint16 = 0x48 // I2CAddr1 is the default I2C device address for the S-5851A.
	I2CAddr2 uint16 = 0x49 // I2CAddr2 I2C device address alternative.
	I2CAddr3 uint16 = 0x4A // I2CAddr3 I2C device address alternative.
	I2CAddr4 uint16 = 0x4b // I2CAddr4 I2C device address alternative.
	I2CAddr5 uint16 = 0x4c // I2CAddr5 I2C device address alternative.
	I2CAddr6 uint16 = 0x4d // I2CAddr6 I2C device address alternative.
	I2CAddr7 uint16 = 0x4e // I2CAddr7 I2C device address alternative.
	I2CAddr8 uint16 = 0x4f // I2CAddr8 I2C device address alternative.

	maximumTemp physic.Temperature = 120*physic.Celsius + (273150 * physic.MilliKelvin) // Maximum anticipated temperature (Upper error detection bound).
	minimumTemp physic.Temperature = -35*physic.Celsius + (273150 * physic.MilliKelvin) // Minimum anticipated temperature (Lower error detection bound).

	regTemp   byte = 0
	regConfig byte = 1

	bitTrigger = 0b10000000 // bitTrigger corresponds to (new temp trigger) bit in config register (1 = temperature acquistion in progress, 0 = Done).
)

var (
	data           = []byte{0, 0} // data is container for I2C sent/received data frames.
	isPtrTemp bool = false        // isPtrTemp indicates whether or not pointer is pointing to temperature readout register.
)

// Opts holds the configuration options.
type Opts struct {
	I2cAddress uint16 // I2C Address of device.
	maxTemp    physic.Temperature
	minTemp    physic.Temperature
}

// DefaultOpts are the recommended default options.
var DefaultOpts = Opts{
	I2cAddress: I2CAddr1,
	maxTemp:    maximumTemp,
	minTemp:    minimumTemp,
}

// Dev is a handle to the device methods.
type Dev struct {
	c    i2c.Dev
	name string
	mu   sync.Mutex // mu inhibits opportunities for simultaneous I2C operations.
}

// NewS5851A creates a new driver for the S-5851A temperature sensor IC.
func NewS5851A(busI2C i2c.Bus, opts *Opts) (*Dev, error) {
	return &Dev{
		c:    i2c.Dev{Bus: busI2C, Addr: opts.I2cAddress},
		name: "S-5851A",
	}, nil
}

// Shutdown controls sleep mode for minimal power consumption.
// isSleep = True = Shutdown = Low power use / one-shot temp conversion mode.
// isSleep = False = Wake device = continuous temperature conversion mode.
func (d *Dev) Shutdown(isSleep bool) (err error) {
	// Lock device to inhibit attempts at multiple simultaneous read/writes.
	d.mu.Lock()
	defer d.mu.Unlock()

	isPtrTemp = false // No longer pointing to temperature read register.

	if isSleep {
		// Set shutdown bit in config register (sleep mode).
		if err := d.c.Tx([]byte{regConfig, 1}, nil); err != nil {
			return errors.New("failed to shutdown " + d.name + ". " + err.Error())
		}
	} else {
		// Clear shutdown bit (awake mode) in config register.
		if err := d.c.Tx([]byte{regConfig, 0}, nil); err != nil {
			return errors.New("failed to wake from sleep " + d.name + ". " + err.Error())
		}
	}
	return nil
}

// Trigger a single (one-shot) temperature acquisition.
// Note function automatically enters sleep mode for low power consumption.
func (d *Dev) OneShotTrigger() (err error) {
	// Lock device to inhibit attempts at multiple simultaneous read/writes.
	d.mu.Lock()
	defer d.mu.Unlock()

	isPtrTemp = false // No longer pointing to temperature read register.

	// Trigger a one shot temperature data acquistion and then sleep.
	if err = d.c.Tx([]byte{regConfig, 129}, nil); err != nil {
		return errors.New("failed to trigger a one-shot temperature acquisition. " + err.Error())
	}
	return nil
}

// Monitor for completion of one-shot temperature acquisition.
// Returns True when acquisition is complete and readout available.
// Returns False when Device is still busy with temperature acquisition.
func (d *Dev) IsTempReady() (isDone bool, err error) {
	// Lock device to inhibit attempts at multiple simultaneous read/writes.
	d.mu.Lock()
	defer d.mu.Unlock()

	confData := []byte{0} // confData = I2C data read from device.

	isPtrTemp = false // No longer pointing to temperature read register.

	if err = d.c.Tx([]byte{regConfig}, nil); err != nil {
		return false, errors.New("failed poll (write to " + d.name + ") for Done signal. " + err.Error())
	}

	if err = d.c.Tx(nil, confData); err != nil {
		return false, errors.New("failed poll (read from " + d.name + ") for Done signal. " + err.Error())
	}

	if confData[0]&bitTrigger != 0 {
		return false, nil // Indicate temperature aquisition still in progress.
	} else {
		return true, nil // Indicate temperature aquisition is complete (ready to read).
	}
}

// ReadTemperature reads temperature from S-5851A device.
func (d *Dev) ReadTemperature() (temperature physic.Temperature, err error) {
	// Lock device to inhibit attempts at multiple simultaneous read/writes.
	d.mu.Lock()
	defer d.mu.Unlock()

	// On either startup or changed pointer status,
	// initialize device pointer to point to (temperature read) register.
	if !isPtrTemp {
		isPtrTemp = true
		if err = d.c.Tx([]byte{regTemp}, nil); err != nil {
			return 0, errors.New("failed to initialize (" + d.name + ") pointer register. " + err.Error())
		}
	}

	// Read raw (temperature) data from device.
	if err := d.c.Tx(nil, data); err != nil {
		return 0, errors.New("failed to read temp from " + d.name + ". " + err.Error())
	}

	// Normalize temperature measurement.
	tempval := uint16(data[0])<<4 | uint16(data[1])>>4

	if int8(data[0]) < 0 {
		tempval = (^tempval) + 1 // If temperature is negative, apply 2's complement.
		tempval = tempval & 4095 // Suppress leading bits inverted by 2's complement as we only need first 12 bits.
		// (-0.0625°C per unit value x 10^6 °nC). Kelvin Offset (273.15°K).
		temperature = (physic.Temperature(tempval) * -62500000) + 273150000000
	} else {
		// (0.0625°C per unit value x 10^6 °nC). Kelvin Offset (273.15°K).
		temperature = (physic.Temperature(tempval) * 62500000) + 273150000000
	}

	// Provide some boundary (sanity) checks on temperature read.
	if temperature > DefaultOpts.maxTemp || (temperature < DefaultOpts.minTemp) {
		err = errors.New("temperature out of bounds. " + d.name + " measured " + temperature.String())
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
