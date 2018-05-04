package parsergen

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type ParserHolder struct {
	Data map[string]map[string]int64
}

func CopyMap(m map[string]int64) map[string]int64 {
	ret := make(map[string]int64)
	for k, v := range m {
		ret[k] = v
	}
	return ret
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

func NewParserHolder() *ParserHolder {
	ph := ParserHolder{}
	ph.Data = make(map[string]map[string]int64)
	return &ph
}

func HandleInputStream(rd io.Reader, data chan<- string) {
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

func ParseRecord(record string) (*ParserHolder, error) {
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
