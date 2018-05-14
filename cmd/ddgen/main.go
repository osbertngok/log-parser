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
	latestPH, err := parsergen.FromReader(rd)
	if err != nil {
		panic(err)
	}

	mapJSON, err := json.MarshalIndent(latestPH.Data, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", mapJSON)
}
