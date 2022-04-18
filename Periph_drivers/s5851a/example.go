// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package s5851a_test

import (
	"fmt"
	"log"
	"time"

	"main.go/example/s5851a" // Revise (driver) path to suit.
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
		fmt.Println("Error: Failed to start I²C communications")
		log.Fatalln(err)
	}
	defer bus.Close()

	// Create handle to S-5851A Ambient temperature measurement device.
	devTemp, err := s5851a.NewS5851A(bus, &s5851a.DefaultOpts)
	if err != nil {
		fmt.Println("Error: Failed to establish comms with device.")
		log.Fatalln(err)
	}

	// Exit Sleep Mode (start continuous temperature measurements).
	// Only necessary here, if device in sleep mode on software startup.
	err = devTemp.Shutdown(false)
	if err != nil {
		log.Fatalln(err)
	}

	// Continuous temperature measurements.
	for start := time.Now(); time.Since(start) < 3*time.Second; { // Run test for finite time period.
		temperature, err := devTemp.ReadTemperature()
		if err != nil {
			log.Fatalln(err)
		} else {
			fmt.Printf("Ambient Temperature is %+v\n", temperature)
		}

		time.Sleep(500 * time.Millisecond) // Arbitrary delay between reads.
	}

	// Demonstrate one-shot temperature conversions in shutdown mode.
	// In this mode the device is asleep most of the time.
	// Power consumption is kept to an absolute minimum.
	// Only wakes on command for new temperature measurement, then re-sleeps.

	// Enter (low power) Sleep Mode.
	// Continuous temperature acquisitions stop.
	// Temperature reads will be valid for last temperature acquisition made.
	err = devTemp.Shutdown(true)
	if err != nil {
		log.Fatalln(err)
	}

	for start := time.Now(); time.Since(start) < time.Second; { // Run demo for finite period.
		// Trigger one-shot temperature acquisition.
		// Note one shot trigger will automatically enter sleep mode once done.
		err = devTemp.OneShotTrigger()
		if err != nil {
			log.Fatalln(err)
		}

		tmrStart := time.Now() // Calculate conversion time for purpose of demo.
		isDone := false

		// Wait for temperature conversion to complete.
		// Note first conversion takes less time,
		// as it started prior to initiating shutdown mode.
		for !isDone {
			isDone, err = devTemp.IsTempReady()
			if err != nil {
				log.Fatalln(err)
			}
		}
		elapsed := time.Since(tmrStart)

		// Read latest temperature data acquisition.
		temperature, err := devTemp.ReadTemperature()
		if err != nil {
			log.Fatalln(err)
		} else {
			fmt.Printf("Ambient Temperature is %+v\n", temperature)
			fmt.Printf("Took %s for device to acquire new temperature.\n", elapsed)
		}
	}
}
