package headers

import (
	"errors"
	"slices"
	"strings"
	"unicode"
)

var specialCharacter = []rune{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	str := string(data)
	idx := strings.Index(str, "\r\n")

	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	supposedHeader := str[:idx]

	n += len(supposedHeader) + 2

	colonIdx := strings.Index(supposedHeader, ":")

	if colonIdx != -1 {
		if !unicode.IsLetter(rune(supposedHeader[colonIdx-1])) {
			return 0, false, errors.New("invalid header")
		}

		key := supposedHeader[:colonIdx]
		key = strings.TrimLeft(key, " ")

		for _, character := range key {

			if !unicode.IsLetter(character) &&
				!unicode.IsDigit(character) &&
				!slices.Contains(specialCharacter, character) {
				return 0, false, errors.New("invalid header")
			}
		}

		key = strings.ToLower(key)

		value := supposedHeader[colonIdx+1:]
		value = strings.TrimSpace(value)

		if curr, exists := h[key]; exists {
			value = curr + ", " + value
		}

		h[key] = value

	}

	return n, false, nil
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)

	_, exists := h[key]

	if !exists {
		return "", false
	}

	return h[key], true
}

func (h Headers) Set(key string, value string) {
	h[key] = value
}

func NewHeaders() Headers {
	return make(Headers)
}
