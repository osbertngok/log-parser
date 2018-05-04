package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/osbertngok/log-parser/parsergen"
	"io"
	"os"
)

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

	go parsergen.HandleInputStream(rd, data)

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
