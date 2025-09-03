package utils

import "regexp"

var digitOnlyRegex = regexp.MustCompile(`[^0-9]`)

// CleanDocument remove todos os símbolos e deixa apenas números
func CleanDocument(doc string) string {
	return digitOnlyRegex.ReplaceAllString(doc, "")
}
