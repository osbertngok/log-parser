// +build ignore

package models

import (
	"github.com/osbertngok/log-parser/parsergen"
	"os"
)

// go:generate go run gen.go

func main() {
	wf, err := os.Create("./dict.json")
	if err != nil {
		panic(err)
	}
	defer wf.Close()

	rf, err := os.Open("./template.log")
	if err != nil {
		panic(err)
	}
	defer rf.Close()

	ph := parsergen.FromReader(rf)
	wf.WriteString(ph.ToJSON())
}
