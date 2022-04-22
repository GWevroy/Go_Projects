## MAX17040 Periph.io Device Driver

This driver interface is written to work within the [Periph.io Library](https://periph.io/) framework. It provides all the I2C facilities required to communicate with the WNK series of industrial temperature and pressure probes.

Refer here for the [WNK Sensor manufacturer](https://www.wnksensor.com/)

The periph.io compatible device driver provides the following sensor communication features

* Read detected temperature (Degrees C by default, but Degrees Kelvin or Fahrenheit are selectable too)
* Read detected pressure (kPa by default, but other units of measurement available as well)
  
Note the availability of features is dependent on WNK sensor type used. In other words, if the probe measures pressure only, reading temperature will result in an error being returned.