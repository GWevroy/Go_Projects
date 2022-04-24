## WNK Pressure/Temperature Periph.io Device Driver

This driver interface is written to work within the [Periph.io Library](https://periph.io/) framework. It provides all the I2C facilities required to communicate with the WNK series of industrial temperature and pressure probes.

Refer here for the [WNK Sensor manufacturer](https://www.wnksensor.com/)

As per the (I2C serial communications) datasheet provided by the manufacturer, the device driver should be fully compatible with the following models:
WNK805, WNK21, WNK19, WNK811, WNK80mA, WNK8010. It is likely to work with other models too. Consult the manufacturer for further information.

The periph.io compatible device driver provides the following sensor (I2C) communication features

* Read detected temperature (°C by default, but Kelvin or Fahrenheit are selectable too)
* Read detected pressure (kPa by default, but other units of measurement available as well)
  
This driver has been hardware tested with a (WNK805 100kPa) pressure sensor only. Although the facility for temperature sensing has been included, and validated in software unit tests, the temperature sensing function has yet to be tested on physical hardware. It should work as expected, however message me (the author) should you incur any problems in its implementation.

The same constructor is used regardless of probe type (temperature, pressure, or both). If the physical probe is exclusively temperature, simply pass in a value of 0 for all respective pressure arguments in the constructor
Example: **probeT, err := NewSensorWNK(0, 0, 0, &bus, &DefaultOpts)**

For either a pressure probe or combination pressure/temperature probe, include the pressure related arguments in the constructor. In the following example, a probe is specified with a 0-100kPa range, a minimum pressure limit of -20kPa, and a maximum pressure limit of 200kPa. The minimum pressure should not be lower than -24kPa, as this is unneccessary and incalculable.
Example: **probeTP, err := NewSensorWNK(100, -20, 200, &bus, &DefaultOpts)**

It is important that the pressure limits are well outside the expected realistic range of the probe, as they do not serve to detect abnormal pressure levels. Instead they provide a means to detect abnormal probe behaviour (malfunction) and/or noisy communications. This detection is provided as a minimum level of protection by the device driver, however the user may wish to further enhance such protective measures, depending on their application.

In the event of a call being made to one of the peripheral reads (temperature or pressure) where that particular physical hardware is not present, an error will be returned advising the respective register is unreachable.
