package common

import (
	"bytes"
	"math"
	"fmt"
	"strconv"
)

func HandleFloat64(buffer *bytes.Buffer, f float64) {
	floor := math.Floor(f)
	if math.Abs(f - floor) < 0.000001 {
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
