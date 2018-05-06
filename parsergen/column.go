package parsergen

import (
	"strings"
)

type Column struct {
	Index     int64    `json:"index"`
	KeyChains []string `json:"keyChains"`
	ValueType string   `json:"valueType"`
}

type Table struct {
	Data []Column `json:"data"`
}

func NewTable() *Table {
	ret := Table{}
	ret.Data = make([]Column, 0)
	return &ret
}

func (t *Table) DeepCopy() *Table {
	ret := NewTable()
	ret.Data = make([]Column, len(t.Data))
	for index, item := range t.Data {
		keyChains := make([]string, len(item.KeyChains))
		copy(keyChains, item.KeyChains)
		t.Data[index] = Column{
			Index:     item.Index,
			KeyChains: keyChains,
			ValueType: item.ValueType,
		}
	}
	return ret
}

func FromParserHolder(t *Table, ph *ParserHolder) *Table {
	var ret *Table
	if t != nil {
		ret = t.DeepCopy()
	} else {
		ret = NewTable()
	}
	var cursor = int64(len(ret.Data))
	m := make(map[string]int64)

	// Initialize hashmap
	for _, item := range ret.Data {
		m[strings.Join(item.KeyChains, ".")] = item.Index
	}
	for k, v := range ph.Data {
		// Does it exist in map?
		if val, ok := m[k]; ok {
			if ret.Data[val].ValueType != convertToType(v) {
				ret.Data[val].ValueType = "mixed"
			}
		} else {
			ret.Data = append(ret.Data, Column{
				Index:     cursor,
				KeyChains: strings.Split(k, "."),
				ValueType: convertToType(v),
			})
			m[k] = cursor
			cursor++
		}
	}

	return ret
}
