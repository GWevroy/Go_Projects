## S-5851A Periph.io Device Driver

This driver interface is written to work within the [Periph.io Library](https://periph.io/) framework. It provides all the I2C facilities required to communicate with the Ablic S-5851A (temperature sensor) integrated circuit.

Refer here for the [S-5851A Temperature Sensor Datasheet](https://www.ablic.com/en/doc/datasheet/temperature_sensor/S5851A_E.pdf)

The datasheet provides details regarding the various features available, all of which have been implemented in the device driver:

* Read temperature in continous sample mode.
* Read Temperature in one-shot trigger mode for lowest power consumption.
* Control Shutdown mode (Sleep / Awake)

Refer to the example_test.go file for demonstration code and how to utlize the device driver. The code was tested on a Raspberry Pi 4 running Raspberry Pi OS.
