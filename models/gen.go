//+build ignore

package main

import (
	"encoding/json"
	"fmt"
	"github.com/osbertngok/log-parser/parsergen"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const DICT_JSON_FILENAME_SUFFIX = "_dict.json"

func loadTable(t *parsergen.Table, className string) bool {
	filename := strings.ToLower(className) + DICT_JSON_FILENAME_SUFFIX
	fmt.Printf("Loading existing %s...\n", filename)
	// load dict.json if exists
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		// it is fine if the file doesn't exist
		// we simply won't load it
		fmt.Printf("%s not found.\n", filename)
		return false
	}

	fmt.Printf("%s found.\n", filename)
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
	ph, err := parsergen.FromReader(rf)
	return parsergen.FromParserHolder(t, ph), nil
}

func writeToJSONDict(t *parsergen.Table, className string) error {
	var (
		jsonStr []byte
		err     error
		wf      *os.File
	)
	if jsonStr, err = json.MarshalIndent(t, "", "    "); err != nil {
		return err
	}
	if wf, err = os.Create(strings.ToLower(className) + DICT_JSON_FILENAME_SUFFIX); err != nil {
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

func keyChainsToClickhouseFields(keyChains []string) string {
	ret := ""
	for _, item := range keyChains {
		ret += "_" + parsergen.GetGoFieldName(item)
	}
	return ret[1:]
}

func writeToRecordStruct(t *parsergen.Table, className string) error {
	node := t.ToNode()
	var (
		err error
		wf  *os.File
		tpl *template.Template
	)
	if tpl, err = template.New("struct.txt.tmpl").Funcs(template.FuncMap{
		"keyChainsToGoFields": keyChainsToGoFields,
	}).ParseFiles("struct.txt.tmpl"); err != nil {
		return err
	}
	if wf, err = os.Create("./" + strings.ToLower(className) + "/index.go"); err != nil {
		return err
	}
	if err = tpl.Execute(wf, struct {
		ClassName   string
		RecordClass string
		Table       *parsergen.Table
	}{
		className,
		node.ToGoClass("", "    ", className),
		t,
	}); err != nil {
		return err
	}
	return nil
}

func writeToCreateSQLScript(t *parsergen.Table, className string) error {
	var (
		err error
		wf  *os.File
		tpl *template.Template
	)
	if tpl, err = template.New("sql.txt.tmpl").Funcs(template.FuncMap{
		"keyChainsToClickhouseFields": keyChainsToClickhouseFields,
	}).ParseFiles("./sql.txt.tmpl"); err != nil {
		return err
	}
	if wf, err = os.Create("./" + strings.ToLower(className) + "/create.sql"); err != nil {
		return err
	}
	if err = tpl.Execute(wf, struct {
		ClassName string
		Table     *parsergen.Table
	}{
		className,
		t,
	}); err != nil {
		return err
	}
	return nil
}

func main() {
	// for _, className := range []string{"OrderCycle", "PLC"} {
	for _, className := range []string{"OrderCycle"} {
		lowerCaseClassName := strings.ToLower(className)
		var (
			t     *parsergen.Table = nil
			files []string
			err   error
		)
		loadTable(t, className)
		logFileLocation := "../data/" + lowerCaseClassName + "/*.log"

		if files, err = filepath.Glob(logFileLocation); err != nil {
			panic(err)
		}

		for _, filename := range files {
			if t, err = enrich(filename, t); err != nil {
				panic(err)
			}
		}

		if t == nil {
			panic(fmt.Errorf("no log files under %s", logFileLocation))
		}

		if err = writeToJSONDict(t, className); err != nil {
			panic(err)
		}

		if err = writeToRecordStruct(t, className); err != nil {
			panic(err)
		}

		if err = writeToCreateSQLScript(t, className); err != nil {
			panic(err)
		}
	}
}
