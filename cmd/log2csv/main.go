package main

import (
	"flag"
	"fmt"
	"github.com/osbertngok/log-parser/models"
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
	records := models.FromReader(rd)

	for _, record := range records {
		fmt.Printf("%s\n", record.ToCSV())
	}
}
