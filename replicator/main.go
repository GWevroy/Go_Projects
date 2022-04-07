// Auto file updater by Graham Ward
//
// Graham Ward
// Version 0.1.0
//
// This program includes code for secure SSH communication between local client and remote server with authentication keys
// Follow the instructions below to manually generate and deploy the required keys
//
// Resources:
// https://linuxize.com/post/how-to-set-up-ssh-keys-on-ubuntu-20-04/
// https://serverfault.com/questions/309171/possible-to-change-email-address-in-keypair
// https://skarlso.github.io/2019/02/17/go-ssh-with-host-key-verification/
//
// 1. Follow instructions found at resource web page on creation of id_rsa private and public keys
//    Note: email address is simply a comment line that takes no action.
//          I set no password when prompted.
//    ssh-keygen -t rsa -b 4096 -C "your_email@domain.com"
//
// 2. Confirm keys now exist (should see id_rsa private and id_rsa.pub public key files):
//    ls ~/.ssh/id_*
//
// 3. Copy keys over from local host to remote server or Raspberry Pi etc.:
//    ssh-copy-id remote_username@server_ip_address

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-fsnotify/fsnotify" // Library used to detect file changes. Download files manually and extract in github repository
	"github.com/pkg/sftp"

	"github.com/maruel/interrupt" // Package ensures graceful exit and proper shutdown of deferred code on CTRL+C Exit

	"golang.org/x/crypto/ssh"
	kh "golang.org/x/crypto/ssh/knownhosts"

	"encoding/json"
)

const (
	configFile = "replicator.conf" // Name of configuration file (can include a path in filename too, but not mandatory)
)

// variables as loaded from config file
var (
	srcUser     string // Local user (required for authentication purposes)
	EventPeriod int64  // Minimum time period (ms) between file updates (eliminates multiple event triggers for what is a single event)
	srcFilePath string // Path to source file that is to be replicated
	srcFileName string // Source filename
	tgtFilePath string // Target file path
	tgtFileName string // Target filename (can be left blank if same as source filename)
	tgtAddress  string // IP Address of target server/Raspberry Pi/etc
	tgtPort     int    // IP Address port for target
	tgtUser     string // Username as required for remote SSH target access

)

type configParams struct {
	SrcUser     string `json:"Local_User"`           // Local user (required for authentication purposes)
	EventPeriod int64  `json:"Min_ms_Update_Period"` // Minimum time in milliseconds that mus transpire between file updates
	SrcFilePath string `json:"Source_Filepath"`      // Path to source file that is to be replicated
	SrcFileName string `json:"Source_Filename"`      // Source filename
	TgtFilePath string `json:"Target_Filepath"`      // Target file path
	TgtFileName string `json:"Target_Filename"`      // Target filename (can be left blank if same as source filename)
	TgtAddress  string `json:"Target_IP_Address"`    // IP Address of target server/Raspberry Pi/etc
	TgtPort     int    `json:"Target_Address_Port"`  // IP Address port for target
	TgtUser     string `json:"Target_Username"`      // Username as required for remote SSH target access
}

var config = configParams{} // Prepare parameter resource (as loaded from configuration file)

// Import configuration parameters from config file
func getConfig() {

	// Does a configuration file exist? If not we need to create a new one complete with default settings
	_, err := os.Stat(configFile)
	if os.IsNotExist(err) {
		fmt.Println("Config file does not exist! Creating new file...")

		// Create new (default) data
		config.SrcUser = "/home/graham"
		config.SrcFilePath = "/home/graham/"
		config.SrcFileName = "myTestFile.txt"
		config.TgtFilePath = "home/graham/Documents/"
		config.TgtAddress = "192.168.1.126"
		config.TgtPort = 22
		config.TgtUser = "graham"
		config.EventPeriod = 10

		file, err := json.MarshalIndent(config, "", " ")
		if err != nil {
			log.Fatal("Error: Default JSON parameter settings for config file are corrupt. Check Code!")
		}
		// Write marshalled parameters to a new configuration file
		err = ioutil.WriteFile(configFile, file, 0644)
		if err != nil {
			log.Fatal("Failed to create new config file. Confirm target drive/volume is ok. Confirm directory permissions.")
		}
		log.Fatal("New file successfully created. Edit " + configFile + " to include source and target file!") // Terminate program at this point as user has to update config file before proceeding
	}

	// File confirmed to exist. Safe to now attempt to load parameters from file
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal("ERROR: Failed to load configuration file (" + configFile + ").\nTry deleting corrupt file and restart program to create a new default file.")
	}

	err = json.Unmarshal([]byte(file), &config)
	if err != nil {
		log.Fatal("ERROR: Failed to load parameters from " + configFile + " file. \nConfirm JSON parameters are not corrupt. \nTry deleting corrupt file and restart program to create a new default file.")
	}

	// If no target filename is provided, default to using the same name as source
	if config.TgtFileName == "" {
		fmt.Println("Target filename is not set in config file. Target File will be saved with the same name as source (" + config.TgtFilePath + config.SrcFileName + ")")
		config.TgtFileName = config.SrcFileName
	}
	fmt.Printf("Local user (required for authentication purposes):\n  " + config.SrcUser)
	fmt.Printf("\nSource to copy:\n  " + config.SrcFilePath + config.SrcFileName)
	fmt.Printf("\nTarget machine address:\n  " + config.TgtUser + "@" + config.TgtAddress + ":" + strconv.Itoa(config.TgtPort))
	fmt.Printf("\nTarget Directory/File:\n  " + config.TgtFilePath + config.TgtFileName + "\n")
	fmt.Printf("Minimum interval between update events (inhibits multiple events for what is in fact a single event):\n  " + strconv.FormatInt(config.EventPeriod, 10) + "ms\n")
}

func main() {

	getConfig() // Import settings from config file

	// Prepare to remotely connect using SSH. Acquire local SSH key
	key, err := ioutil.ReadFile(config.SrcUser + "/.ssh/id_rsa")
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	hostKeyCallback, err := kh.New(config.SrcUser + "/.ssh/known_hosts")
	if err != nil {
		log.Fatal("could not create hostkeycallback function: ", err)
	}

	configSSH := &ssh.ClientConfig{
		User: config.TgtUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer), // Could add password check here for additional security.
		},
		HostKeyCallback: hostKeyCallback,
	}
	// Connect to the remote server and perform the SSH handshake.
	sshConnection, err := ssh.Dial("tcp", config.TgtAddress+":"+strconv.Itoa(config.TgtPort), configSSH)
	if err != nil {
		fmt.Println("Error: Ensure you have copied keys over from client to server (possible cause).")
		log.Fatalf("unable to connect: %v", err)
	}
	defer sshConnection.Close()

	// create new SFTP client
	client, err := sftp.NewClient(sshConnection)
	if err != nil {
		log.Fatalf("ERROR: Failed to apply SFTP protocol layer to SSH connection: %v", err)
	}
	defer client.Close()

	interrupt.HandleCtrlC() // Manages abrupt exits of program to ensure deferred code is always run prior to process kills

	fmt.Println("SSH (transport layer) Communication channel opened.")
	fmt.Println("SFTP session successfully established. Hit CTRL+C to exit at any time.")

	// create a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer watcher.Close()

	quit := make(chan bool) // Listen for any quit command from main loop
	go func() {

		lastUpdateTime := time.Now() // Prepare to measure time duration between events (to eliminate multiple events for what is in fact a single event)
		eventCounter := 0            // Tally any and all updates
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				{
					tmrNow := time.Now()
					duration := int64(tmrNow.Sub(lastUpdateTime) / time.Millisecond) // Calculate how much time has passed (in ms) since the last valid event

					if event.Op&fsnotify.Write != fsnotify.Write {
						log.Fatal("ERROR: Source file appears to have been deleted or no longer accessible: ", event.Name)
					} else if duration > config.EventPeriod {
						eventCounter++ // This is a valid update event. Increment tally of update events

						dstFile, err := client.Create(config.TgtFilePath + config.TgtFileName)
						if err != nil {
							fmt.Println("Fatal Error: Check destination filepath")
							log.Fatal(err)

						}

						// create source file
						srcFile, err := os.Open(config.SrcFilePath + config.SrcFileName)
						if err != nil {
							fmt.Println("Fatal Error: Check source filepath and file")
							log.Fatal(err)
						}

						// copy source file to destination file
						bytes, err := io.Copy(dstFile, srcFile)
						if err != nil {
							fmt.Println("ERROR: Failed to update file!")
							log.Fatal(err)
						}
						fmt.Printf("%v Updates thus far. %d bytes copied\n", eventCounter, bytes)

						dstFile.Close() // Close source and destination files until next event
						srcFile.Close()
						lastUpdateTime = time.Now() // Update timer to evalute next update interval
					}

				}

			// watch for errors
			case err := <-watcher.Errors:
				fmt.Println("ERROR: Event Watcher failed: ", err)

				// Prepare to exit at a moments notice (implemented to handle CTRL+C exits of Main loop)
			case <-quit:
				return
			}
		}
	}()

	// out of the box fsnotify can watch a single file, or directory. Code in Go routine though only copies files at this time
	if err := watcher.Add(config.SrcFilePath + config.SrcFileName); err != nil {
		log.Fatal("ERROR: Local file/directory watcher:", err)
	}

	for {
		// Monitor for abrupt program exit
		if interrupt.IsSet() {
			quit <- true // Terminate Go routine
			fmt.Println("\nCTRL+C hit! Program exiting gracefully")
			break
		}
		time.Sleep(20 * time.Millisecond) // Pause before polling again (else system CPU resource is hammered for no reason)
	}
	fmt.Printf("Program exit at %v\n", time.Now().Format("03:04:05"))
}
