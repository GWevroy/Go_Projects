// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Note this library is likely equally well suited for use with MAX17041,
// but untested for that device. The library could also be easily refactored
// for use with the rest of the Maxim MAX1704x series of devices.

package max17040

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/physic"
)

// Device driver constants.
const (
	// Device (16 bit) Registers.
	VCell   byte = 0x02 // Cell Voltage register. Read only. Address 02h - 03h.
	SoC     byte = 0x04 // State Of Charge register. Read only. Address 04h - 05h.
	Mode    byte = 0x06 // Mode register. Write only. Responsible for managing special commands. Address 06h - 07h.
	Version byte = 0x08 // Register returns IC Version. Read only. Address 08h - 09h.
	RComp   byte = 0x0c // Battery Compensation register. Read/Write. Adjust IC performance. Address 0Ch - 0Dh.
	Command byte = 0xfe // Register sends special commands. Write only. Address FEh - FFh.

	MaxV physic.ElectricPotential = 4250 * physic.MilliVolt // Absolute maximum anticipated voltage (Upper error detection bound).
	MinV physic.ElectricPotential = 2200 * physic.MilliVolt // Absolute minimum anticipated voltage (Lower error detection bound).

	MinRCOMP int = 10000 // Minimum RCOMP value permissible.
	MaxRCOMP int = 55000 // Maximum RCOMP value permissible.

	I2CAddr    uint16 = 0x36   // I2CAddr is the default I2C address for the MAX17040/1.
	QuickStart uint16 = 0x4000 // QuickStart is the default (Quick-Start) value used for the Mode Register.
	RcompCal   uint16 = 0x9700 // RcompCal is the default calibration value for the RCOMP Register.
	CMDpor     uint16 = 0x0054 // CMDpor is the default (Power On Reset) command value for the Command Register.
)

var data = []byte{0, 0} // data is used to format I2C sent/received data frames.

// Opts holds the configuration options.
type Opts struct {
	I2cAddress uint16                   // I2C Address of device.
	Mode       uint16                   // Mode register setting.
	RCOMP      uint16                   // RCOMP register setting.
	Command    uint16                   // Command register setting.
	MinRComp   int                      // Minimum RCOMP calibration value.
	MaxRComp   int                      // Maximum RCOMP calibration value.
	MaxVolts   physic.ElectricPotential // Maximum voltage threshold.
	MinVolts   physic.ElectricPotential // Minimum voltage threshold.
}

// DefaultOpts are the recommended default options.
var DefaultOpts = Opts{
	I2cAddress: I2CAddr,
	Mode:       QuickStart,
	RCOMP:      RcompCal,
	Command:    CMDpor,
	MinRComp:   MinRCOMP,
	MaxRComp:   MaxRCOMP,
	MaxVolts:   MaxV,
	MinVolts:   MinV,
}

// Dev is a handle to the MAX17040/1 device.
type Dev struct {
	c    i2c.Dev
	name string
	mu   sync.Mutex // mu inhibits opportunities for simultaneous I2C operations.
}

// NewMAX17040 creates a new driver for the MAX17040 Li-ion Fuel Gauge IC.
func NewMAX17040(busI2C i2c.Bus, opts *Opts) (*Dev, error) {
	return &Dev{
		c:    i2c.Dev{Bus: busI2C, Addr: opts.I2cAddress},
		name: "MAX17040",
	}, nil
}

// Reset function performs a reset similar to a power cycle.
func (d *Dev) Reset() (err error) {
	// Lock device to inhibit attempts at multiple simultaneous read/writes.
	d.mu.Lock()
	defer d.mu.Unlock()

	bufReset := make([]byte, 2)
	binary.BigEndian.PutUint16(bufReset, DefaultOpts.Command) // Encode reset command.
	bufReset = append([]byte{Command}, bufReset...)           // Prepend bufReset slice with the Command register address.

	if err := d.c.Tx(bufReset, nil); err != nil {
		err = fmt.Errorf("failed to reset ("+d.name+") device. %w", err) // prepend error with additional debug information.
		return err
	}
	return nil
}

// QuickStart reboots the (State of Charge) alogrithm without resetting device
func (d *Dev) QuickStart() (err error) {
	// Lock device to inhibit attempts at multiple simultaneous read/writes.
	d.mu.Lock()
	defer d.mu.Unlock()

	bufReset := make([]byte, 2)
	binary.BigEndian.PutUint16(bufReset, DefaultOpts.Mode) // Encode Quickstart mode (Default).
	bufReset = append([]byte{Mode}, bufReset...)           // Prepend bufReset slice with the Mode register address.

	if err := d.c.Tx(bufReset, nil); err != nil {
		err = fmt.Errorf("failed to assert mode for ("+d.name+") device. %w", err) // prepend error with additional debug information.
		return err
	}
	return nil
}

//  SetRCOMP writes new RCOMP calibration value.
// Note value is volatile and will revert to default on power cycle or reset.
func (d *Dev) SetRCOMP(rcompVal int) (err error) {
	// Lock device to inhibit attempts at multiple simultaneous read/writes.
	d.mu.Lock()
	defer d.mu.Unlock()

	// Perform a (bounds) sanity check on the proposed new value for RCOMP.
	if (rcompVal > DefaultOpts.MaxRComp) || (rcompVal < DefaultOpts.MinRComp) {
		err = errors.New("proposed new RCOMP value (" + strconv.Itoa(rcompVal) + ") is out of bounds. Value unchanged")
		return err
	}

	bufRcomp := make([]byte, 2)
	binary.BigEndian.PutUint16(bufRcomp, uint16(rcompVal)) // Encode Reset command.
	bufRcomp = append([]byte{RComp}, bufRcomp...)          // Prepend bufReset slice with the RCOMP register address.

	if err := d.c.Tx(bufRcomp, nil); err != nil {
		fmt.Printf("failed to update R-Compensation for "+d.name+". %v\n", err)
		return err
	}
	return nil
}

// GetRCOMP reads RCOMP calibration value.
func (d *Dev) GetRCOMP() (rcompVal int, err error) {
	// Lock device to inhibit attempts at multiple simultaneous read/writes.
	d.mu.Lock()
	defer d.mu.Unlock()

	// Read raw (RCOMP) data from device.
	if err := d.c.Tx([]byte{RComp}, data); err != nil {
		err = fmt.Errorf("failed to fetch "+d.name+" RCOMP value. %w", err) // prepend error with additional debug information.
		return 0, err
	}

	rcompVal = int(data[0])<<8 | int(data[1]) // Calculate RCOMP value based on raw data from device.

	// Perform a sanity check on the value recieved.
	if (rcompVal > DefaultOpts.MaxRComp) || (rcompVal < DefaultOpts.MinRComp) {
		err = errors.New("received RCOMP value out of bounds. Detected value: " + strconv.Itoa(rcompVal))
		return 0, err
	}
	return rcompVal, nil
}

// GetVersion reads Version information.
func (d *Dev) GetVersion() (version string, err error) {
	// Lock device to inhibit attempts at multiple simultaneous read/writes.
	d.mu.Lock()
	defer d.mu.Unlock()

	// Convert byte (ASCII literal) to string.
	byteTostring := func(val byte) string {
		num := int(val)
		if num < 10 {
			return "0" + strconv.Itoa(num) // Provide some leading padding as required.
		}
		return strconv.Itoa(num)
	}

	// Read raw (version information) data from device.
	if err := d.c.Tx([]byte{Version}, data); err != nil {
		err = fmt.Errorf("failed to fetch "+d.name+" Version. %w", err) // prepend error with additional debug information.
		return "", err
	}

	version = byteTostring(data[0]) + "/" + byteTostring(data[1]) // Return formatted string value.
	return version, nil
}

// ReadCellVoltage reads Cell or Battery DC Voltage.
func (d *Dev) ReadCellVoltage() (voltage physic.ElectricPotential, err error) {
	// Lock device to inhibit attempts at multiple simultaneous read/writes.
	d.mu.Lock()
	defer d.mu.Unlock()

	// Read raw (cell voltage) data from device.
	if err := d.c.Tx([]byte{VCell}, data); err != nil {
		return 0, err
	}

	// Calculate DC voltage.
	// (1 250 000 coefficient is 1.25mV per unit value x 10^6 nV).
	voltage = (physic.ElectricPotential(data[0])<<4 | physic.ElectricPotential(data[1])>>4) * 1250000

	// Provide some boundary (sanity) checks on voltage read.
	if (voltage > DefaultOpts.MaxVolts) || (voltage < DefaultOpts.MinVolts) {
		err = errors.New("failed to read voltage from " + d.name)
		return 0, err
	}
	return voltage, nil
}

// ReadSoC reads State of Charge (Returns Percentage).
func (d *Dev) ReadSoC() (SOCpercent float32, err error) {
	// Lock device to inhibit attempts at multiple simultaneous read/writes.
	d.mu.Lock()
	defer d.mu.Unlock()

	// Specify SoC (State of Charge) register to be read from the device.
	if err := d.c.Tx([]byte{SoC}, data); err != nil {
		return 0, err
	}

	SOCpercent = float32(uint16(data[0])<<8|uint16(data[1])) / 256 //(Calculate State of Charge %).

	// Provide some boundary (sanity) checks on SoC read.
	if (SOCpercent > 100) || (SOCpercent < 0) {
		err = errors.New("failed to read State of Charge from " + d.name)
		return 0, err
	}
	return SOCpercent, nil
}

// String implements conn.Resource.
func (d *Dev) String() string {
	return d.name
}

// Halt implements conn.Resource.
func (d *Dev) Halt() error {
	return nil
}
