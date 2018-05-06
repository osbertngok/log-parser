//+build ignore

package main

import (
	"encoding/json"
	"github.com/osbertngok/log-parser/parsergen"
	"io/ioutil"
	"os"
	"path/filepath"
	"fmt"
)

const DICT_JSON_FILENAME = "./dict.json"
const LOG_FILES_PATTERN = "../data/*.log"

func loadTable(t *parsergen.Table) bool {
	fmt.Printf("Loading existing %s...\n", DICT_JSON_FILENAME)
	// load dict.json if exists
	raw, err := ioutil.ReadFile(DICT_JSON_FILENAME)
	if err != nil {
		// it is fine if the file doesn't exist
		// we simply won't load it
		fmt.Printf("%s not found.\n", DICT_JSON_FILENAME)
		return false
	}

	fmt.Printf("%s found.\n", DICT_JSON_FILENAME)
	err = json.Unmarshal(raw, t)
	return err == nil
}

func enrich(filename string, t *parsergen.Table) *parsergen.Table {
	fmt.Printf("Loading %s...\n", filename)
	rf, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer rf.Close()
	ph := parsergen.FromReader(rf)
	return parsergen.FromParserHolder(t, ph)
}

func main() {
	var t *parsergen.Table = nil
	loadTable(t)

	wf, err := os.Create(DICT_JSON_FILENAME)
	if err != nil {
		panic(err)
	}
	defer wf.Close()

	files, err := filepath.Glob(LOG_FILES_PATTERN)
	for _, filename := range files {
		t = enrich(filename, t)
	}

	jsonStr, err := json.MarshalIndent(t, "", "    ")
	if err != nil {
		panic(err)
	}
	wf.WriteString(string(jsonStr))
}
