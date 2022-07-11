---
title: "MMBASIC Data Parser"
---

# MMBASIC Data Parser
#### by Graham Ward
#### Programming language: Go version 1.18.1

## Overview
This little program is a tool to facilitate MMBASIC code development where DATA lines are concerned. If you have calculated data, routinely revised data, or simply a lot of data that you wish to efficiently incorporate into your MMBASIC program (DATA lines), then this tool may be of value to you. This saves you the time and effort of manually typing out the data, as well as the potential for human error to sneak in. It permits you to take data from a .CSV file (for example, from a text file or spreadsheet), import and parse it into an acceptable MMBASIC .BAS data format.

The data parser takes in data from a specified file (Default: DataSource.csv) and parses the data in accordance with the configuration settings specified in **data_parser.conf**. The output data is then saved to an output file (Default: Output.bas)

When the program is first run, it loads user preferences from the configuration file (**data_parser.conf**) as found in the same directory. If the configuration file is not found, a new one is automatically generated, along with default settings. The configuration file is in a JSON format, and may be edited to set the following parameters:

* Data_is_text - True/False (String/Numerical)
* Column_Count - Maximum character count before next Data line is automatically declared. Default 80.
* Delimiter    - comma, space, tab etc. Default: comma
* Source_File  - Specify (optional) path and filename. Default: DataSource.csv
* Target_File  - Specify (optional) path and filename. Default: Output.bas

For numerical data, remember to set the (**Data_is_text**) parameter to false. This will result in data points being packed without being enclosed in quotation marks. Conversely for text strings, set the parameter to true.

Text data will have any leading/trailing whitespace automatically trimmed off.

The delimiter separated data can either be in a single column, single row, or a blend of both. The only criteria the program cares about is the delimiter itself.

Refer to the example target file (**Output.bas**) that was produced by the source file (**DataSource.csv**) that have been included for your reference.

The **Column_Count** parameter ensures that no **DATA** line shall be longer than the specified number of characters. This includes the **DATA** command, any field separating commas, quotation marks etc. As soon as a given field will potentially extend past this boundary point, the program will automatically generate a new **DATA** line.

## Useage

1. On your PC (running either Windows or Linux) download the applicable binary or executable file (executables have been created for both operating systems. Sorry MAC users, you will have to compile from source which is included here)  
2. Ensure the configuration file is appropriately setup. If the file does not exist (IE you have yet to run the program for the first time) then execute the program (ignoring input/output requirements). With the default file created, edit it to suit.
3. Confirm your Source (.CSV) file is prepared, and execute the program
4. Check your Target (.BAS) file and confirm results are as desired.
5. At this point, you have the option of either copying the Target file over to your MMBASIC machine's SD card, or use use a terminal client like **PuTTY** or **Screen** to copy/paste the data lines directly from the target file to the MMBASIC editor.

I hope you find this little program useful. Any suggestions or comments, let me know.

