## MAX17040 Periph.io Device Driver

This driver interface is written to work within the [Periph.io Library](https://periph.io/) framework. It provides all the I2C facilities required to communicate with the MAXIM MAX17040 (fuel gauge) integrated circuit. While it has not been tested on the MAXIM17041 it will in all likelihood run without issue driving this chip too.

Refer here for the [MAX17040/MAX17041 Datasheet](https://datasheets.maximintegrated.com/en/ds/MAX17040-MAX17041.pdf)

The datasheet provides details regarding the various registers available.
The device driver provides the following I2C facilities for the MAX17040 fuel gauge IC:

* Read cell or battery Voltage (returned as type physic.ElectricPotential)
* Read State of Charge (Percentage, returned as Float32)
* Power On Reset command (Identical to a power cycle)
* Quick-start command for rebooting the (State of Charge) algorithm without initiating a full device reset or power cycle
* Ability to read IC Version information
* Read/Write RCOMP calibration value

Refer to the example_test.go file for demonstration code and how to utlize the device driver. This code was tested on a Raspberry Pi 4 running Raspberry Pi OS, with a [geekworm X728 DC UPS](https://wiki.geekworm.com/X728) HAT that incorporates the Maxim IC.
