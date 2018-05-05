package parsergen

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func CopyMap(m map[string]int64) map[string]int64 {
	ret := make(map[string]int64)
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

type ParserHolder struct {
	Data map[string]map[string]int64
}

func (ph *ParserHolder) AddPair(key string, innerType string) {
	if _, ok := ph.Data[key]; !ok {
		ph.Data[key] = make(map[string]int64)
	}
	m := ph.Data[key]
	m[innerType]++
}

func (ph *ParserHolder) DeepCopy() *ParserHolder {
	ret := NewParserHolder()
	for key, value := range ph.Data {
		ret.Data[key] = CopyMap(value)
	}
	return ret
}

func (ph1 *ParserHolder) Append(ph2 *ParserHolder) {
	for k, v := range ph2.Data {
		if v1, ok := ph1.Data[k]; ok {
			for k2, v2 := range v {
				v1[k2] += v2
			}
		} else {
			ph1.Data[k] = CopyMap(v)
		}
	}
}

func convertToType(m map[string]int64) string {
	keys := make([]string, 0)
	for key, _ := range m {
		keys = append(keys, key)
	}
	if len(keys) == 1 {
		return keys[0]
	} else {
		return "mixed"
	}
}

func (ph *ParserHolder) ToJSON() string {
	ret := "{\n\t\"data\": ["
	keyLength := len(ph.Data)
	ct := 0
	for k, v := range ph.Data {
		ret += "{\n"
		ret += fmt.Sprintf("\t\t\"index\":%d,\n", ct)
		ret += fmt.Sprintf("\t\t\"keyChains\": [%s],\n", strings.Split(k, "."))
		ret += fmt.Sprintf("\t\t\"valueType\": \"%s\",\n", convertToType(v))
		ret += "\t}"
		if ct != keyLength-1 {
			ret += ", "
		}
		ct++
	}
	ret += "\n"
	return ret
}

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

func NewParserHolder() *ParserHolder {
	ph := ParserHolder{}
	ph.Data = make(map[string]map[string]int64)
	return &ph
}

func FromReader(rd io.Reader) *ParserHolder {
	var ret *ParserHolder = nil
	data := make(chan string, 1000)

	go handleInputStream(rd, data)

	for record := range data {
		if record != "" {
			newPH, err := ParseRecord(record)
			if err != nil {
				fmt.Printf("%s, %v\n", record, err)
				continue
			}
			if ret == nil {
				ret = newPH
			} else {
				ret.Append(newPH)
			}
		}
	}
	return ret
}
