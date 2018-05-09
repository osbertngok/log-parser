package models

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

const RFC3339Micro = "2006-01-02T15:04:05.000000-0700"

func NewRecord() *Record {
	return &Record{}
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
	err = json.Unmarshal([]byte(JSONstr), ret)
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
