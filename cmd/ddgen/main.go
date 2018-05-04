package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/osbertngok/log-parser/parsergen"
	"io"
	"os"
)

func handleInputStream(rd io.Reader, data chan<- string) {
	reader := bufio.NewReader(rd)
	var err error = nil
	for {
		var subline []byte
		var line []byte
		isPrefix := true
		ct := 0

		// read until reaches end of line (!isPrefix),
		// or reaches end of file (err)
		for isPrefix && err == nil {
			ct++
			// read until buffer is full (isPrefix),
			// or reaches end of line (!isPrefix),
			// or reaches end of file (err)
			subline, isPrefix, err = reader.ReadLine()
			line = append(line, subline...)
		}
		data <- string(line)
		// if reaches end of file (or other error)
		// break the loop
		// and close the channel
		if err != nil {
			break
		}
	}
	close(data)
}

func main() {

	var filename string
	flag.StringVar(&filename, "f", "", "file to read from")
	var rd io.Reader = os.Stdin

	// If filename is specified, override stdin
	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		rd = file
	}

	var latestPH *parsergen.ParserHolder = nil
	data := make(chan string, 1000)

	go handleInputStream(rd, data)

	for record := range data {
		if record != "" {
			newPH, err := parsergen.ParseRecord(record)
			if err != nil {
				fmt.Printf("%s, %v\n", record, err)
				continue
			}
			if latestPH == nil {
				latestPH = newPH
			} else {
				latestPH.Append(newPH)
			}
		}
	}
	mapJSON, err := json.MarshalIndent(latestPH.Data, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", mapJSON)
}
