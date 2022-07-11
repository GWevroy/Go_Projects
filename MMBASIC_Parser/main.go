// MMBASIC Data Parser

// Copyright (c) 2022 Graham Ward

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"main.go/appConfig" // Application Configuration file management (opened on start of app, and all parameters configured from here)
)

// WriteToFile creates a new text file with final MMBASIC code output
func WriteToFile(text string) error {
	file, err := os.Create(appConfig.Conf.TargetFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(text)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	appConfig.LoadConfig() // Load configuration / calibration values from configuration file.

	file, err := os.Open(appConfig.Conf.SourceFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var sb strings.Builder // String builder is the most efficient method for string concatenation
	var words []string
	scanner := bufio.NewScanner(file) // Open text file, ready for scanning each line of file
	for scanner.Scan() {
		sb.WriteString(scanner.Text()) // Concatenate all lines into a single string
	}

	if err := scanner.Err(); err != nil { // Trap any file reading errors
		log.Fatal(err)
	}
	words = append(words, strings.Split(sb.String(), appConfig.Conf.Delimiter)...) // Create word array using "," delimiter

	// Trim any leading and trailing whitespace
	for j := range words {
		words[j] = strings.TrimSpace(words[j])
	}

	// build MMBASIC data-line of code
	var lineMMBASIC strings.Builder
	isFirstWord := true
	lineLen := appConfig.Conf.ColCount + 5 // initiate line length calculation. Arbitrary value greater than MMBASIC max column length

	cntComp := 1 // Default to adjust compensation for determining column count (including comma only)
	if appConfig.Conf.Stringify {
		cntComp = 3 // Adjust compensation for determining column count (including comma and quotation marks)
	}

	for j := 0; j < len(words); j++ {
		// Add next word in queue (up to 80 columns)
		if lineLen+len(words[j])+cntComp <= appConfig.Conf.ColCount {
			lineLen = lineLen + len(words[j]) + cntComp // Upate line length tracker + 3 for quotation marks and comma
			if isFirstWord {
				isFirstWord = false
			} else {
				lineMMBASIC.WriteString(",")
			}
			if appConfig.Conf.Stringify {
				lineMMBASIC.WriteString("\"") // Add (start) quotation mark as required
			}
			lineMMBASIC.WriteString(words[j]) // Add next data value
			if appConfig.Conf.Stringify {
				lineMMBASIC.WriteString("\"") // Add (end) quotation mark as required
			}
		} else {
			if j == 0 {
				lineMMBASIC.WriteString("Data ") // Mitigate newline character at start of file

			} else {
				lineMMBASIC.WriteString("\nData ") // Max column count exceeded. Start new line.
			}
			if appConfig.Conf.Stringify {
				lineMMBASIC.WriteString("\"") // Add (start) quotation mark as required
			}
			lineMMBASIC.WriteString(words[j]) // Add next data value
			if appConfig.Conf.Stringify {
				lineMMBASIC.WriteString("\"") // Add (end) quotation mark as required
			}
			lineMMBASIC.WriteString(",") // Add MMBASIC Data delimiter
			isFirstWord = true
			lineLen = 4 + cntComp + len(words[j]) // reset line length calculation. 7 = "data " + (apostrophe x2)
		}
	}
	// Write string to output file
	err = WriteToFile(lineMMBASIC.String())
	if err != nil {
		log.Fatalf("Failed to write to file. %v", err)
	}
	fmt.Println("MMBASIC file created/updated successfully.")
}
