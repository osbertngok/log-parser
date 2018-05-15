package common

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
	"time"
)

type IRow struct {
	EventDate    time.Time
	EventTime    time.Time
	Microsecond  int64
	ControllerNo int64
	Bitmap0      uint64
	Bitmap1      uint64
	Bitmap2      uint64
	Bitmap3      uint64
	Bitmap4      uint64
	Bitmap5      uint64
	Bitmap6      uint64
	Bitmap7      uint64
	Bitmap8      uint64
	Bitmap9      uint64
}

func HandleFloat64(buffer *bytes.Buffer, f float64) {
	floor := math.Floor(f)
	if math.Abs(f-floor) < 0.000001 {
		buffer.WriteString(fmt.Sprintf("%s,", strconv.FormatFloat(f, 'f', 0, 64)))
	} else {
		buffer.WriteString(fmt.Sprintf("%s,", strconv.FormatFloat(f, 'f', 6, 64)))
	}
}

func HandleString(buffer *bytes.Buffer, s string) {
	buffer.WriteString(fmt.Sprintf("\"%s\",", s))
}

func MuteAsString(buffer *bytes.Buffer) {
	buffer.WriteString("\"\",")
}

func HandleBool(buffer *bytes.Buffer, b bool) {
	buffer.WriteString(fmt.Sprintf("%s,", boolToIntString(b)))
}

func boolToIntString(b bool) string {
	if b {
		return "1"
	} else {
		return "0"
	}
}

func HandleInputStream(rd io.Reader, data chan<- string) {
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
