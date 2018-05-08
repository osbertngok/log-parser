//+build ignore

package main

import (
	"encoding/json"
	"fmt"
	"github.com/osbertngok/log-parser/parsergen"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

const DICT_JSON_FILENAME = "./dict.json"
const LOG_FILES_PATTERN = "../data/*.log"
const RECORD_GO_FILE = "./record.go"
const RECORD_TEMPLATE_FILENAME = "./record.txt.tmpl"

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

func enrich(filename string, t *parsergen.Table) (*parsergen.Table, error) {
	fmt.Printf("Loading %s...\n", filename)
	rf, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer rf.Close()
	ph := parsergen.FromReader(rf)
	return parsergen.FromParserHolder(t, ph), nil
}

func writeToJSONDict(t *parsergen.Table, filename string) error {
	var (
		jsonStr []byte
		err     error
		wf      *os.File
	)
	if jsonStr, err = json.MarshalIndent(t, "", "    "); err != nil {
		return err
	}
	if wf, err = os.Create(filename); err != nil {
		return err
	}
	defer wf.Close()
	if _, err := wf.WriteString(string(jsonStr)); err != nil {
		return err
	}
	return nil
}

func keyChainsToGoFields(keyChains []string) string {
	ret := ""
	for _, item := range keyChains {
		ret += "." + parsergen.GetGoFieldName(item)
	}
	return ret
}

func writeToRecordStruct(t *parsergen.Table, filename string) error {
	node := t.ToNode()
	var (
		err error
		wf  *os.File
		tpl *template.Template
	)
	if tpl, err = template.New("record.txt.tmpl").Funcs(template.FuncMap{
		"keyChainsToGoFields": keyChainsToGoFields,
	}).ParseFiles(RECORD_TEMPLATE_FILENAME); err != nil {
		return err
	}
	if wf, err = os.Create(filename); err != nil {
		return err
	}
	if err = tpl.Execute(wf, struct {
		RecordClass string
		Table       *parsergen.Table
	}{
		node.ToGoClass("", "    "),
		t,
	}); err != nil {
		return err
	}
	return nil
}

func main() {
	var (
		t     *parsergen.Table = nil
		files []string
		err   error
	)
	loadTable(t)

	if files, err = filepath.Glob(LOG_FILES_PATTERN); err != nil {
		panic(err)
	}
	for _, filename := range files {
		if t, err = enrich(filename, t); err != nil {
			panic(err)
		}
	}

	if err = writeToJSONDict(t, DICT_JSON_FILENAME); err != nil {
		panic(err)
	}

	if err = writeToRecordStruct(t, RECORD_GO_FILE); err != nil {
		panic(err)
	}

}
