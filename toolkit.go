package main

func SubStr(str string, start, end int) string {
	if len(str) == 0 {
		return ""
	}
	if end >= len(str) {
		end = len(str) - 1
	}
	return str[start:end]
}
