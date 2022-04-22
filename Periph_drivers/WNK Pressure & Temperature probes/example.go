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
	// Min Pressure (fatal error if exceeded) = -35kPa
	// Max Pressure (fatal error if exceeded) = 150kPa
	devTemp, err := wnk.NewSensorWNK(100, -35, 150, bus, &wnk.DefaultOpts)
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

// // *******************
// // * Pressure Sensor *
// // *******************
// func pressure(botPressureI2C *i2c.I2C) {
// 	var snPressure float64

// 	kPaDat, _, err := botPressureI2C.ReadRegBytes(0x06, 3)
// 	if err != nil {
// 		if appConfig.Conf.IsVerbose {
// 			fmt.Println("Error: Failed to communicate with pressure sensor")
// 			fmt.Println(err)
// 		}
// 		return
// 	} else if appConfig.Conf.IsVerbose {
// 		fmt.Println("Info: Read request made successfully to pressure sensor")
// 	}

// 	tmpkPa := uint32(kPaDat[0])<<16 | uint32(kPaDat[1])<<8 | uint32(kPaDat[2]) // Big Endian (left shift MSB) integer created from bytes.
// 	if (tmpkPa & 0x800000) != 0 {
// 		snPressure = float64(tmpkPa) - 16777216.0 // Adjust for negative (sensed) pressure if applicable.
// 	} else {
// 		snPressure = float64(tmpkPa)
// 	}
// 	snPressure = 3.3 * snPressure / 8388608
// 	snPressure = kpaRange * (snPressure - 0.5) / 2

// 	snPressure = (snPressure * appConfig.Conf.Pres1_Gain) + appConfig.Conf.Pres1_Offset // Apply calibration settings.

// 	if !appConfig.Conf.IsMute {
// 		fmt.Printf("Pressure is %.3fkPa \n", snPressure)
// 	}
// 	botPressure.Set(snPressure) // Save tank bottom pressure sensor reading to database.
// }
