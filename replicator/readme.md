---
title: "Auto file updater"
---
# Auto File Replicator
#### by Graham Ward
#### Programming language: Go version 1.13.8
#### Compiled/run on Ubuntu 20.04 Linux 64 Bit
#### Version 1.0.0


## Overview
The purpose of this program is to monitor a specified local file for any changes. When a change or write occurs to the file in question, an event is immediately triggered. The event then copies the source file to the designated target destination. The target destination can be any computer, server, Raspberry Pi etc. that is network accessible from the local (source) computer.

Multiple instances of the program may be run concurrently in order to monitor and update different files. You'll have to create different directories for each copy of the program to accomodate the configurations (the IP address and port obviously remain the same for each config file, without incurring any SSH conflicts). Note I could have added some additional facility to monitor an entire directory (recursively) but for my requirement this wasn't necessary. I also felt that for larger projects it may be prudent to not be copying all the files that are not changing, and instead focus on only replicating the files that do change.

This program was written to help me make software development on the Raspberry Pi simpler. IE although I do regularly cross compile code, there are times when what I am developing simply will not cross compile and has to be compiled natively on the Raspberry Pi itself. For this reason I developed this application that will monitor my source file as I write code, and as soon as I save it, automatically copy it over (using SFTP over an SSH connection with full authentication/encryption). I can then turn to the Raspberry Pi and compile the code just copied over to it. The simple yet tedius task of manually copying and pasting the file over a mapped network drive (or frequent use of an application like Filezilla) is thus mitigated.

The program should have no problem replicating files of any kind across a network. It's use is not limited to text files.

The code is written in Golang and serves as a working example of how to get SSH communications up and running, along with SFTP, using Go. It utilizes a watcher that relies on an OS API to monitor for any file changes that may occur. This minimizes program overhead.
the code is tested/verified to compile and run in Ubuntu Linux. It should compile and run in Windows too, however I have not verified this yet.

In order to ensure secure communications, SSH keys need to be manually generated and copied over to the target device as well. The process is relatively simple though.
Follow the instructions below (or see the resource link for more information) to manually generate and deploy the required keys

   1. Verify whether or not the keys exist already. You should see **id_rsa** private and **id_rsa.pub** public key files listed when you execute:
   
      **ls ~/.ssh/id_***

   2. If the keys do not exist, create SSH **id_rsa** private and public keys.
         Note: email address is simply a comment line entered into the generated key file and has no action or effect.
               Refer to resource below for more information.
               I set no password when prompted, however you can, should you deem it necessary.

      **ssh-keygen -t rsa -b 4096 -C "your_email@domain.com"**

   3. Confirm keys now exist (should see **id_rsa** private and **id_rsa.pub** public key files):
   
      **ls ~/.ssh/id_***

   4. Copy keys over from local host to remote server or Raspberry Pi etc.:
   
      **ssh-copy-id remote_username@server_ip_address**

## Operation
1. The application is run from the command line. Either compile from source or use the ready made (Linux 64Bit) executable (replicator). To execute simply go to the installation directory and type in:
   
   **./replicator**

2. When the program is first started it will look for a configuration file. If one is not found it will auto-generate a new one with default settings.
   You will need to edit the default config (JSON) file to suit your particular requirements (source and tartget directories and files).
   The target file name is optional. If ommited the program will default to using the source file name.

3. Once that is done you are free to restart the program which will then display your current configuration. If there are any errors at any point the program will immediately abort.

4 Once it gets to the point of successful connection, try making a change to your source file and confirm it is immediately copied to the target destination. Feedback is provided in the terminal window as well, where you will see a tally of the number of updates and how many bytes were copied over.

5. As the instructions state, hit CTRL+C to quit once done.

## Resources:

https://linuxize.com/post/how-to-set-up-ssh-keys-on-ubuntu-20-04/

https://serverfault.com/questions/309171/possible-to-change-email-address-in-keypair

https://skarlso.github.io/2019/02/17/go-ssh-with-host-key-verification/

