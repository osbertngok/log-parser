//+build ignore

package main

import (
	"github.com/osbertngok/log-parser/parsergen"
	"os"
	"encoding/json"
	"io/ioutil"
)

func loadTable(t *parsergen.Table) bool {
	// load dict.json if exists
	raw, err := ioutil.ReadFile("./dict.json")
	if err != nil {
		// it is fine if the file doesn't exist
		// we simply won't load it
		return false
	}

	err = json.Unmarshal(raw, t)
	return err == nil
}

func main() {
	var initT *parsergen.Table = nil
	loadTable(initT)

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
