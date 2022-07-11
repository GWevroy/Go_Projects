package appConfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func LoadConfig() {
	//--------------------------------------
	//encoding / Marshalling
	//--------------------------------------
	jsonConfData, _ := json.Marshal(Conf)

	// Open Sensor Array Configuration file and fetch calibration values
	configFile, err := os.OpenFile(configFileName, os.O_RDONLY, 0600) //0600 sets permissions same as CHMOD
	if err != nil {
		if os.IsNotExist(err) {
			newConfigFile, err := os.OpenFile(configFileName, os.O_RDWR|os.O_CREATE, 0600) //0600 sets permissions same as CHMOD
			if err != nil {
				fmt.Println("Error: Attempt to create a new configuration file failed.")
				log.Fatal(err)
			}

			_, err = newConfigFile.Write(jsonConfData) // Store default parameters to newly made file
			if err != nil {
				fmt.Println("Error: failed to store parameter data to newly created Configuration file")
				log.Fatal(err)
			}

			err = newConfigFile.Close() // Appropriately manage closure of writable file
			if err != nil {
				fmt.Println("Error: failed to save newly created config file. Check storage medium!")
				log.Fatal(err)
			}

			fmt.Print("Info: Configuration file does not exist. New file successfully created. Default paramater values assumed.")
		} else {
			fmt.Println("Error: failed to open Configuration file. check for corrupt media or permissions.")
			log.Fatal(err)
		}

	} else {
		defer configFile.Close() // File is read only, so need to manage any likely errors

		//decode (unmarsall) parameter data from file
		byteValue, _ := ioutil.ReadAll(configFile)
		err = json.Unmarshal(byteValue, &Conf)
		if err != nil {
			log.Fatal(err)
		}
	}
	println()
}
