package parsergen

import (
	"sort"
	"strings"
	"fmt"
)

type Column struct {
	Index     int64    `json:"index"`
	KeyChains []string `json:"keyChains"`
	ValueType string   `json:"valueType"`
}

type Table struct {
	Data []Column `json:"data"`
}

type Node struct {
	Index              int64
	LogName            string
	DatabaseColumnName string
	GoType             string
	GoFieldName        string
	Children           map[string]*Node
}

var rootFields = []*Node{
	{
		Index:              0,
		LogName:            "",
		DatabaseColumnName: "EventTime",
		GoType:             "time.Time",
		GoFieldName:        "EventTime",
		Children:           nil,
	},
	{
		Index:              1,
		LogName:            "",
		DatabaseColumnName: "Microsecond",
		GoType:             "int64",
		GoFieldName:        "Microsecond",
		Children:           nil,
	},
	{
		Index:              2,
		LogName:            "",
		DatabaseColumnName: "ControllerNo",
		GoType:             "int64",
		GoFieldName:        "ControllerNo",
		Children:           nil,
	},
}

const RESERVED_COLUMNS_NO = 3

func (n *Node) ToGoClass(prefix string, tab string) string {
	tags := ""
	if n.LogName != "" {
		tags += fmt.Sprintf(" json:\"%s\"", n.LogName)
	}
	if n.Index >= 0 {
		tags += fmt.Sprintf(" index:%d", n.Index)
	}
	ret := prefix
	if n.GoFieldName == "" {
		ret += "type Record "
	} else {
		ret += prefix + n.GoFieldName + " "
	}
	if n.GoType != "" {
		ret += n.GoType
	} else {
		ret += "struct {\n"
		if n.GoFieldName == "" {
			for _, node := range rootFields {
				ret += node.ToGoClass(prefix+tab, tab)
			}
		}
		keys := make([]string, 0)
		for key, _ := range n.Children {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			v, _ := n.Children[key]
			ret += v.ToGoClass(prefix+tab, tab)
		}
		ret += prefix + "}"
	}
	if tags != "" {
		ret += fmt.Sprintf("`%s`", strings.TrimLeft(tags, " "))
	}
	ret += "\n"
	return ret
}

func getDatabaseColumnName(keyChains []string) string {
	return strings.Join(keyChains, ".")
}

func getGoFieldName(key string) string {
	ret := strings.ToUpper(key[:1]) + key[1:]
	for _, t := range []string{"[", "]", ":"} {
		ret = strings.Replace(ret, t, "_", -1)
	}
	return ret
}

func getGoType(valueType string) string {
	switch valueType {
	case "bool":
		return "bool"
	case "float64":
		return "float64"
	case "string":
		return "string"
	case "mixed":
		return "string"
	default:
		return "string"
	}
}

func (t *Table) ToNode() *Node {
	root := &Node{
		Index:              -1,
		LogName:            "",
		DatabaseColumnName: "",
		GoFieldName:        "",
		GoType:             "",
		Children:           nil,
	}
	for _, column := range t.Data {
		parentNode := root
		for index, key := range column.KeyChains {
			if parentNode.Children == nil {
				parentNode.Children = make(map[string]*Node)
			}
			var (
				currentNode *Node
				ok          bool
			)
			if currentNode, ok = parentNode.Children[key]; !ok {
				if index == len(column.KeyChains)-1 {
					// leaf
					currentNode = &Node{
						Index:              column.Index,
						LogName:            key,
						DatabaseColumnName: getDatabaseColumnName(column.KeyChains),
						GoFieldName:        getGoFieldName(key),
						GoType:             getGoType(column.ValueType),
						Children:           nil,
					}
				} else {
					currentNode = &Node{
						Index:              -1,
						LogName:            key,
						DatabaseColumnName: "",
						GoFieldName:        getGoFieldName(key),
						GoType:             "",
						Children:           nil,
					}
				}
				parentNode.Children[key] = currentNode
			}
			parentNode = currentNode
		}
	}
	return root
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
		ret.Data[index] = Column{
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
		m[strings.Join(item.KeyChains, ".")] = item.Index - RESERVED_COLUMNS_NO
	}
	for k, v := range ph.Data {
		// Does it exist in map?
		if val, ok := m[k]; ok {
			if ret.Data[val].ValueType != convertToType(v) {
				ret.Data[val].ValueType = "mixed"
			}
		} else {
			ret.Data = append(ret.Data, Column{
				Index:     cursor + RESERVED_COLUMNS_NO,
				KeyChains: strings.Split(k, "."),
				ValueType: convertToType(v),
			})
			m[k] = cursor
			cursor++
		}
	}

	return ret
}
