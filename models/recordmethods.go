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
	"strconv"
)

const RFC3339Micro = "2006-01-02T15:04:05.000000-0700"

func NewRecord() *Record {
	return &Record{}
}


func subParseString(f interface{}, keyChains []string, r *Record, rv reflect.Value) error {
	m := f.(map[string]interface{})
	if m == nil {
		return errors.New("not an object")
	}
	pv := reflect.Indirect(rv)
	if pv.Kind() != reflect.Struct {
		return fmt.Errorf("%v is not a struct", keyChains)
	}
	for k, v := range m {
		field := pv.FieldByName(parsergen.GetGoFieldName(k))
		structField, ok := pv.Type().FieldByName(parsergen.GetGoFieldName(k))
		if !ok {
			return fmt.Errorf("field %s not found", k)
		}
		if field.IsValid() {
			switch v.(type) {
			case string:
				if field.Type().String() == "string" {
					field.SetString(v.(string))
					intStr := structField.Tag.Get("index")
					index, err := strconv.ParseInt(intStr, 10, 64)
					if err == nil && index >= 0 {
						r.MarkBitmap(uint64(index))
					}
				} else {
					fmt.Printf("%s, type mismatch", field.Type().String())
					// handle mismatch field
				}
			case float64:
				if field.Type().String() == "float64" {
					field.SetFloat(v.(float64))
					intStr := structField.Tag.Get("index")
					index, err := strconv.ParseInt(intStr, 10, 64)
					if err == nil && index >= 0  {
						r.MarkBitmap(uint64(index))
					}
				} else {
					fmt.Printf("%s, type mismatch", field.Type().String())
					// handle mismatch field
				}
			case bool:
				if field.Type().String() == "bool" {
					field.SetBool(v.(bool))
					intStr := structField.Tag.Get("index")
					index, err := strconv.ParseInt(intStr, 10, 64)
					if err == nil && index >= 0  {
						r.MarkBitmap(uint64(index))
					}
				} else {
					fmt.Printf("%s, type mismatch", field.Type().String())
					// handle mismatch field
				}
			case []interface{}:
				if field.Type().String() == "string" {
					field.SetString(fmt.Sprintf("%v", v))
					intStr := structField.Tag.Get("index")
					index, err := strconv.ParseInt(intStr, 10, 64)
					if err == nil && index >= 0  {
						r.MarkBitmap(uint64(index))
					}
				} else {
					fmt.Printf("%s, type mismatch", field.Type().String())
					// handle mismatch field
				}
			case map[string]interface{}:
				if field.CanAddr() {
					subParseString(v, append(keyChains, k), r, reflect.ValueOf(field.Addr().Interface()))
				} else {
					return fmt.Errorf("Cannot Addr for key %s", k)
				}
			case nil:
				fmt.Printf("nil for %s\n", k)

			default:
				fmt.Printf("unknown type: %T", v)
				// handle unknown field
			}
		} else {
			// handle unknown field
		}
	}
	return nil
}

func ParseString(log string, controllerNo int64, tz string) (*Record, error) {
	ret := NewRecord()

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, err
	}

	// Remove timestamp
	i := strings.Index(log, ",")
	if i == -1 {
		return nil, fmt.Errorf("%s does not contain comma", log)
	}
	JSONstr := log[i+1:]
	var f interface{}
	err = json.Unmarshal([]byte(JSONstr), &f)
	if err != nil {
		return nil, err
	}
	m := f.(map[string]interface{})
	subParseString(m, []string{}, ret, reflect.ValueOf(ret))
	if err != nil {
		return nil, err
	}
	ret.ControllerNo = controllerNo
	timestampString := log[:i]
	timestamp, err := time.Parse(RFC3339Micro, timestampString)
	if err != nil {
		return nil, err
	}
	eventDateString := log[:10]
	eventDate, err := time.ParseInLocation("2006-01-02", eventDateString, loc)
	if err != nil {
		return nil, err
	}
	ret.EventDate = eventDate
	ret.Microsecond = int64(timestamp.Nanosecond()) / 1000
	ret.EventTime = timestamp.Add(time.Duration(-1*timestamp.Nanosecond()) * time.Nanosecond)
	return ret, nil
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

func FromReader(rd io.Reader, controllerNo int64, tz string) []*Record {
	var ret = make([]*Record, 0)
	data := make(chan string, 1000)

	go handleInputStream(rd, data)

	for log := range data {
		if log != "" {
			newRecord, err := ParseString(log, controllerNo, tz)
			if err != nil {
				fmt.Printf("%s, %v\n", log, err)
				continue
			}
			ret = append(ret, newRecord)
		}
	}
	return ret
}
