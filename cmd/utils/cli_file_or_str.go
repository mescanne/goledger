package utils

import (
	"fmt"
	"io/ioutil"
	"strings"
)

const FILE_PREFIX = "file:"

func GetFileOrStr(str string) (string, error) {
	if !strings.HasPrefix(str, FILE_PREFIX) {
		return str, nil
	}

	fname := str[len(FILE_PREFIX):len(str)]

	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", fname, err)
	}

	return string(b), nil
}
