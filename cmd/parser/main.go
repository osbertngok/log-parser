package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type ParserHolder struct {
	data map[string]map[string]int64
}

func CopyMap(m map[string]int64) map[string]int64 {
	ret := make(map[string]int64)
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

func (ph *ParserHolder) AddPair(key string, innerType string) {
	if _, ok := ph.data[key]; !ok {
		ph.data[key] = make(map[string]int64)
	}
	m := ph.data[key]
	m[innerType]++
}

func (ph *ParserHolder) DeepCopy() *ParserHolder {
	ret := NewParserHolder()
	for key, value := range ph.data {
		ret.data[key] = CopyMap(value)
	}
	return ret
}

func (ph1 *ParserHolder) Append(ph2 *ParserHolder) {
	for k, v := range ph2.data {
		if v1, ok := ph1.data[k]; ok {
			for k2, v2 := range v {
				v1[k2] += v2
			}
		} else {
			ph1.data[k] = CopyMap(v)
		}
	}
}

func NewParserHolder() *ParserHolder {
	ph := ParserHolder{}
	ph.data = make(map[string]map[string]int64)
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

func subParseRecord(f interface{}, prefix string, ph *ParserHolder) error {
	m := f.(map[string]interface{})
	if m == nil {
		return errors.New("not an object")
	}
	for k, v := range m {
		switch v.(type) {
		case string:
			ph.AddPair(prefix+k, "string")
		case float64:
			ph.AddPair(prefix+k, "float64")
		case bool:
			ph.AddPair(prefix+k, "bool")
		case []interface{}:
			ph.AddPair(prefix+k, "array")
		case map[string]interface{}:
			subParseRecord(v, prefix+k+".", ph)
		default:
			ph.AddPair(prefix+k, fmt.Sprintf("unknown(%T)", v))
		}
	}
	return nil
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

	subParseRecord(m, "", ph)

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
	mapJSON, err := json.MarshalIndent(latestPH.data, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", mapJSON)
}
