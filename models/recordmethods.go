package models

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/osbertngok/log-parser/parsergen"
	"io"
	"reflect"
	"strings"
	"time"
)

const RFC3339Micro = "2006-01-02T15:04:05.000000-0700"

func NewRecord() *Record {
	return &Record{}
}

func subParseString(f interface{}, keyChains []string, record interface{}) error {
	m := f.(map[string]interface{})
	if m == nil {
		return errors.New("not an object")
	}
	for k, v := range m {
		field := reflect.Indirect(reflect.ValueOf(record)).FieldByName(parsergen.GetGoFieldName(k))
		if field.IsValid() {
			switch v.(type) {
			case string:
				if field.Type().String() == "string" {
					field.SetString(v.(string))
				} else {
					// handle mismatch field
				}
			case float64:
				if field.Type().String() == "float64" {
					field.SetFloat(v.(float64))
				} else {
					// handle mismatch field
				}
			case bool:
				if field.Type().String() == "bool" {
					field.SetBool(v.(bool))
				} else {
					// handle mismatch field
				}
			case []interface{}:
				if field.Type().String() == "string" {
					field.SetString(fmt.Sprintf("%v", v))
				} else {
					// handle mismatch field
				}
			case map[string]interface{}:
				subParseString(v, append(keyChains, k), record)
			default:
				// handle unknown field
			}
		} else {
			// handle unknown field
		}

	}
	return nil
}
func ParseString(log string) (*Record, error) {
	record := NewRecord()
	// Remove timestamp
	i := strings.Index(log, ",")
	if i == -1 {
		return nil, fmt.Errorf("%s does not contain comma", log)
	}
	timestampString := log[:i]
	timestamp, err := time.Parse(RFC3339Micro, timestampString)
	if err != nil {
		return nil, err
	}
	record.Microsecond = timestamp.UnixNano() / 1000
	record.EventTime = timestamp.Add(time.Duration(-1 * timestamp.UnixNano()))
	JSONstr := log[i+1:]
	var f interface{}
	err = json.Unmarshal([]byte(JSONstr), &f)
	if err != nil {
		return nil, err
	}
	m := f.(map[string]interface{})
	subParseString(m, []string{}, record)
	return record, nil
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

func FromReader(rd io.Reader) []*Record {
	var ret = make([]*Record, 0)
	data := make(chan string, 1000)

	go handleInputStream(rd, data)

	for record := range data {
		if record != "" {
			newRecord, err := ParseString(record)
			if err != nil {
				fmt.Printf("%s, %v\n", record, err)
				continue
			}
			ret = append(ret, newRecord)
		}
	}
	return ret
}
