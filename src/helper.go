package main

import (
	"strings"
	"strconv"
)

func atoi(s string) uint64 {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	result, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return result
}