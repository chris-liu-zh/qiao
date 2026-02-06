package tools

import (
	"errors"
	"unicode"
)

// CamelCaseToUdnderscore 驼峰单词转下划线单词
func CamelCaseToUdnderscore(s string) string {
	var output []rune
	for i, r := range s {
		if i == 0 {
			output = append(output, unicode.ToLower(r))
		} else {
			if unicode.IsUpper(r) {
				output = append(output, '_')
			}

			output = append(output, unicode.ToLower(r))
		}
	}
	return string(output)
}

func DeferErr(src *error, f func() error) {
	if err := f(); err != nil {
		if *src == nil {
			*src = err
			return
		}
		*src = errors.Join(*src, err)
	}
}
