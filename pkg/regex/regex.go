package my_regex

import (
	"fmt"
	"strings"
)

func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	fmt.Println("GetStringInBetween s:", s)
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return
	}
	e = s + e
	return str[s:e]
}
