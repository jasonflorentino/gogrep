package idxablstr

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type IndexableString []string

func (is IndexableString) CharAt(index int) string {
	if index >= len(is) {
		panic(fmt.Sprintf("Error: index %d is outside len %d", index, len(is)))
	}
	if index < 0 {
		return is[index+len(is)]
	}
	return is[index]
}

func (is IndexableString) String() string {
	return strings.Join(is, "")
}

func FromBytes(bytes []byte) IndexableString {
	newIs := IndexableString{}
	for offset := 0; offset < len(bytes); {
		rune, runeWidth := utf8.DecodeRune(bytes[offset:])
		newIs = append(newIs, string(rune))
		offset += runeWidth
	}
	return newIs
}

func FromString(str string) IndexableString {
	newIs := IndexableString{}
	for _, c := range str {
		newIs = append(newIs, string(c))
	}
	return newIs
}
