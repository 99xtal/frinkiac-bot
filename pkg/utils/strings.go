package utils

import "strings"

func SquishText(s string, width int) string {
	words := strings.Split(s, " ")
	newString := ""
	currentLine := ""
	for _, w := range words {
		if len(currentLine)+len(w) > width {
			newString += currentLine + "\n"
			currentLine = ""
		}
		currentLine += " " + w
	}
	newString += currentLine
	return newString
}
