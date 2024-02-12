package main

import (
	"os"
	"strings"
)

type names map[string]string

func getModNames() (names, error) {
	modNames := make(map[string]string)
	htmlContent, err := os.ReadFile("modlist.html")
	if err != nil {
		return nil, err
	}

	htmlText := string(htmlContent)

	startTag := "<a "
	endTag := "</a>"
	startIndex := 0

	for startIndex != -1 {
		startIndex = strings.Index(htmlText[startIndex:], startTag) + startIndex
		endIndex := strings.Index(htmlText[startIndex:], endTag) + startIndex
		a := htmlText[startIndex:endIndex]
		if len(a) == 0 {
			break
		}
		modNames[strings.Split(a, "/")[4][:6]] = strings.Split(a, ">")[1]
		startIndex = endIndex + 1
	}

	return modNames, nil
}
