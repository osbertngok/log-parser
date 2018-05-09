package main

import (
	"flag"
	"fmt"
	"github.com/osbertngok/log-parser/config"
	"github.com/osbertngok/log-parser/models"
	"io"
	"os"
)

func main() {

	cfg := config.New()

	var filename string
	var controllerNo int64
	flag.StringVar(&filename, "f", "", "file to read from")
	flag.Int64Var(&controllerNo, "c", 0, "controller number")
	flag.Parse()
	var rd io.Reader = os.Stdin

	// If filename is specified, override stdin
	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		rd = file
	}
	records := models.FromReader(rd, controllerNo, cfg.Timezone)

	for _, record := range records {
		fmt.Printf("%s\n", record.ToCSV())
	}
}
