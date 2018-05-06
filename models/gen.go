//+build ignore

package main

import (
	"github.com/osbertngok/log-parser/parsergen"
	"os"
	"encoding/json"
)

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
	t := parsergen.FromParserHolder(nil, ph)
	jsonStr, err := json.MarshalIndent(t, "", "    ")
	if err != nil {
		panic(err)
	}
	wf.WriteString(string(jsonStr))
}
