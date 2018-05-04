package parsergen

func CopyMap(m map[string]int64) map[string]int64 {
	ret := make(map[string]int64)
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

type ParserHolder struct {
	Data map[string]map[string]int64
}

func (ph *ParserHolder) AddPair(key string, innerType string) {
	if _, ok := ph.Data[key]; !ok {
		ph.Data[key] = make(map[string]int64)
	}
	m := ph.Data[key]
	m[innerType]++
}

func (ph *ParserHolder) DeepCopy() *ParserHolder {
	ret := NewParserHolder()
	for key, value := range ph.Data {
		ret.Data[key] = CopyMap(value)
	}
	return ret
}

func (ph1 *ParserHolder) Append(ph2 *ParserHolder) {
	for k, v := range ph2.Data {
		if v1, ok := ph1.Data[k]; ok {
			for k2, v2 := range v {
				v1[k2] += v2
			}
		} else {
			ph1.Data[k] = CopyMap(v)
		}
	}
}

func NewParserHolder() *ParserHolder {
	ph := ParserHolder{}
	ph.Data = make(map[string]map[string]int64)
	return &ph
}
