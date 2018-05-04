package parsergen

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func subParseRecord(f interface{}, prefix string, ph *ParserHolder) error {
	m := f.(map[string]interface{})
	if m == nil {
		return errors.New("not an object")
	}
	for k, v := range m {
		switch v.(type) {
		case string:
			ph.AddPair(prefix+k, "string")
		case float64:
			ph.AddPair(prefix+k, "float64")
		case bool:
			ph.AddPair(prefix+k, "bool")
		case []interface{}:
			ph.AddPair(prefix+k, "array")
		case map[string]interface{}:
			subParseRecord(v, prefix+k+".", ph)
		default:
			ph.AddPair(prefix+k, fmt.Sprintf("unknown(%T)", v))
		}
	}
	return nil
}

func ParseRecord(record string) (*ParserHolder, error) {
	ph := NewParserHolder()
	// Remove timestamp
	i := strings.Index(record, ",")
	if i == -1 {
		return nil, fmt.Errorf("%s does not contain comma", record)
	}
	JSONstr := record[i+1:]
	var f interface{}
	err := json.Unmarshal([]byte(JSONstr), &f)
	if err != nil {
		return nil, err
	}
	m := f.(map[string]interface{})

	subParseRecord(m, "", ph)

	return ph, nil
}
