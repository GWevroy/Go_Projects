package max17040_test

import (
	"fmt"

	"main.go/max17040" // Update with alternative resource location

	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func Example() {
	// Initialize periph.io.
	if _, err := host.Init(); err != nil {
		fmt.Println("Error: Unable to initialize Periph.io library")
		panic(err)
	}

	// Open default I²C bus.
	bus, err := i2creg.Open("")
	if err != nil {
		fmt.Println("Error: Failed to start I²C communications")
		panic(err)
	}
	defer bus.Close()

	// Create handle to MAX17040 device.
	PSUcomms, err := max17040.NewMAX17040(bus, &max17040.DefaultOpts)
	if err != nil {
		fmt.Println("Error: Failed to establish comms with SO2 ADC")
		panic(err)
	}

	// Get Cell DC Voltage measurement.
	dcVoltage, err := PSUcomms.ReadCellVoltage() // Fetch the presently measured DC UPS voltage
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Printf("UPS Voltage: %+v\n", dcVoltage)
		num := float64(dcVoltage) / 1000000000 // Float64 value is in nV. Calc aligns this value to read in Volts.
		fmt.Printf("formatted number = %.2f Volts\n", num)
	}

	// Get Cell State of Charge (Precision Percentage).
	upsSoC, err := PSUcomms.ReadSoC()
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Printf("Cell State of Charge: %.3f%%\n", upsSoC)
	}

	version, err := PSUcomms.GetVersion()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Version of this MAX17040: %v\n", version)
	}

	// Initiate Quickstart Mode (clears algorithm to same state as a power up)
	err = PSUcomms.QuickStart()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Device has entered Quickstart Mode")
	}

	// Verify RCOMP is default value (0x9700 or 38656 decimal).
	compensation, err := PSUcomms.GetRCOMP()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v RCOMP calibration value = %v\n", PSUcomms.String(), compensation)
	}

	// Set or change RCOMP value.
	newRCOMP := 12345
	err = PSUcomms.SetRCOMP(newRCOMP)
	if err != nil {
		fmt.Printf("Failed to set RCOMP. %v\n", err)
	} else {
		fmt.Printf("RCOMP value successfully changed to %v\n", newRCOMP)
	}

	// Verify RCOMP has in fact been changed (Should read the same as above).
	compensation, err = PSUcomms.GetRCOMP()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v RCOMP calibration value = %v\n", PSUcomms.String(), compensation)
	}

	err = PSUcomms.Reset()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Device reset.")
	}

	// Prove Reset has occurred by verifying that RCOMP has reverted to default.
	compensation, err = PSUcomms.GetRCOMP()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v RCOMP calibration value = %v\n", PSUcomms.String(), compensation)
	}

	// Example of changing a default value in device driver.
	max17040.DefaultOpts.MaxRComp = 12000 // Reduce maximum threshold.
	fmt.Printf("Maximum threshold for RCOMP now changed to %v\n", max17040.DefaultOpts.MaxRComp)

	err = PSUcomms.SetRCOMP(newRCOMP) // Attempt to change RCOMP above threshold.
	if err != nil {
		fmt.Printf("Failed to set RCOMP. %v\n", err)
	} else {
		fmt.Printf("RCOMP value successfully changed to %v\n", newRCOMP)
	}
}
