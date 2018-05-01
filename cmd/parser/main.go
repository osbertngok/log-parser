package main

import (
	"bufio"
	"os"
	"fmt"
	"io"
)

func handleInputStream(rd io.Reader, data chan<- string) {
	reader := bufio.NewReader(rd)
	var err error = nil
	for {
		var (isPrefix     bool = true
			subline, line []byte
		)
		ct := 0

		// read until reaches end of line (!isPrefix),
		// or reaches end of file (err)
		for isPrefix && err == nil {
			ct++
			fmt.Printf("subline count: %d\n", ct)
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
	data := make(chan string, 1000)
	go handleInputStream(os.Stdin, data)
}
