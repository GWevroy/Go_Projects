### MAX17040 Periph.io Device Driver


This driver interface is written to work within the <a href="https://periph.io/">Periph.io Library</a> framework. It provides all the I2C facilities required to communicate with the MAXIM MAX17040 (fuel gauge) integrated circuit. While it has not been tested on the MAXIM17041 it will in all likelihood run without issue driving this chip too.

Refer here for the  <a href="https://datasheets.maximintegrated.com/en/ds/MAX17040-MAX17041.pdf">MAX17040/MAX17041 Datasheet</a>.


The datasheet provides details regarding the various registers available.
The device driver provides the following I2C facilities for the MAX17040 fuel gauge IC:

* Read Voltage in nV (returned as type physic.ElectricPotential)
* Read State of Charge (Percentage, returned as Float32)
* Power On Reset command (Identical to a power cycle)
* Ability to read IC Version information
* Change RCOMP calibration value
* Read current RCOMP calibration value

Refer to the example_test.go file for demonstration code and how to utlize the device driver. This code was tested on a Raspberry Pi 4 running Raspberry Pi OS.