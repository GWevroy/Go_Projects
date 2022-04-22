// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wnk_test

import (
	"fmt"
	"log"
	"time"

	"main.go/example/wnk" // Revise (driver) path to suit.
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func Example() {
	// Initialize periph.io.
	if _, err := host.Init(); err != nil {
		fmt.Println("Error: Unable to initialize Periph.io library")
		log.Fatalln(err)
	}

	// Open default I²C bus.
	bus, err := i2creg.Open("")
	if err != nil {
		fmt.Println("Error: Failed to initialize I²C comms peripheral.")
		log.Fatalln(err)
	}
	defer bus.Close()

	// Create handle to WNK Transducer (pressure/temperature) device.
	// Example user settings:
	// Pressure range = 0 - 100kPa = 100kPa
	// Min Pressure (fatal error if exceeded) = -10kPa
	// Max Pressure (fatal error if exceeded) = 150kPa
	devTemp, err := wnk.NewSensorWNK(100, -10, 150, bus, &wnk.DefaultOpts)
	if err != nil {
		fmt.Println("Error: Failed to configure device comms.")
		log.Fatalln(err)
	}

	// Continuous pressure measurements.
	for start := time.Now(); time.Since(start) < 1*time.Second; { // Run test for finite time period.
		sensePressure, err := devTemp.ReadPressure()
		if err != nil {
			fmt.Println("Error: Transducer or I2C malfunction.")
			log.Fatalln(err)
		} else {
			fmt.Printf("Pressure is %+v\n", sensePressure)
		}

		time.Sleep(500 * time.Millisecond) // Arbitrary delay between reads.
	}
}
