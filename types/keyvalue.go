package types

import "strings"

type KeyValue interface {
	Key() string
	Value() string
}

func ReplaceKeyByValue[T KeyValue](variables []T, data string) string {
	if len(variables) == 0 {
		return data
	}

	for _, v := range variables {
		data = strings.ReplaceAll(data, v.Key(), v.Value())
	}

	return data
}
