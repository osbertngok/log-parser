package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"encoding/json"
	"sort"
	"flag"
)

type ParserHolder struct {
	data map[string]string
}

func (ph *ParserHolder) AddPair(key string, innerType string) {
	ph.data[key] = innerType
}

func (ph1 *ParserHolder) Append(ph2 *ParserHolder) (*ParserHolder, error) {
	keys1 := make([]string, 0)
	for k, _ := range ph1.data {
		keys1 = append(keys1, k)
	}

	keys2 := make([]string, 0)
	for k, _ := range ph2.data {
		keys2 = append(keys2, k)
	}
	sort.Strings(keys1)
	sort.Strings(keys2)

	ret := NewParserHolder()
	cursor1 := 0
	cursor2 := 0
	for cursor1 < len(keys1) || cursor2 < len(keys2) {
		var key1, key2, value1, value2 string

		if cursor1 != len(keys1) {
			key1 = keys1[cursor1]
			value1 = ph1.data[key1]
		}
		if cursor2 != len(keys2) {
			key2 = keys2[cursor2]
			value2 = ph2.data[key2]
		}
		if key1 == "" {
			ret.AddPair(key2, value2)
			cursor2++
			continue
		}
		if key2 == "" {
			ret.AddPair(key1, value1)
			cursor1++
			continue
		}
		switch {
		case key1 < key2:
			ret.AddPair(key1, value1)
			cursor1++
		case key1 > key2:
			ret.AddPair(key2, value2)
			cursor2++
		default:
			if ph1.data[key1] != ph2.data[key2] {
				return nil, fmt.Errorf("key %s has different types: %s, %s", key1, value1, value2)
			}
			ret.AddPair(key1, value1)
			cursor1++
			cursor2++
		}
	}
	return ret, nil
}

func NewParserHolder() *ParserHolder {
	ph := ParserHolder{}
	ph.data = make(map[string]string)
	return &ph
}


func handleInputStream(rd io.Reader, data chan<- string) {
	reader := bufio.NewReader(rd)
	var err error = nil
	for {
		var (
			isPrefix      bool = true
			subline, line []byte
		)
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

func parseRecord(record string) (*ParserHolder, error) {
	ph := NewParserHolder()
	// Remove timestamp
	i := strings.Index(record, ",")
	if i == -1 {
		return nil, fmt.Errorf("%s does not contain comma", record)
	}
	JSONstr := record[i+1:]
	var f interface{}
	err := json.Unmarshal([]byte(JSONstr), &f)
	if err != nil {
		return nil, err
	}
	m := f.(map[string]interface{})
	if m == nil {
		return nil, fmt.Errorf("json %s is not an object", JSONstr)
	}
	// First layer
	for k, v := range m {
		switch v.(type) {
		case string:
			ph.AddPair(k, "string")
		case float64:
			ph.AddPair(k, "float64")
		default:
			ph.AddPair(k, "unknown")
		}
	}
	return ph, nil
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

	var latestPH *ParserHolder = nil
	data := make(chan string, 1000)

	go handleInputStream(rd, data)

	for record := range data {
		if record != "" {
			newPH, err := parseRecord(record)
			if err != nil {
				fmt.Printf("%v", err)
				continue
			}
			if latestPH == nil {
				latestPH = newPH
			} else {
				nPH, err := latestPH.Append(newPH)
				if err != nil {
					fmt.Printf("%v", err)
					continue
				}
				latestPH = nPH
			}
		}
	}
	fmt.Printf("%v", *latestPH)
}
