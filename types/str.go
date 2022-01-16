package types

import "strings"

type Str string

func (s Str) String() string {
	return string(s)
}

func (s Str) ToUpper() Str {
	return Str(strings.ToUpper(string(s)))
}

func (s Str) TrimSpace() Str {
	return Str(strings.TrimSpace(string(s)))
}
