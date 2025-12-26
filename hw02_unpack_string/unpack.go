package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	var sb strings.Builder

	r := []rune(s)

	for i := 0; i < len(r); {
		if unicode.IsDigit(r[i]) {
			return "", ErrInvalidString
		}

		char := r[i]
		count := 1
		i++

		if i < len(r) && unicode.IsDigit(r[i]) {
			number, _ := strconv.Atoi(string(r[i]))
			count = number
			i++
		}

		sb.WriteString(strings.Repeat(string(char), count))
	}

	return sb.String(), nil
}
