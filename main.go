package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/ryanfaerman/netctl/internal/events"
)

type (
	NamespaceThingHappened         struct{}
	NamespaceSomethingElseHappened struct{}
)

func splitByCapital(input string) []string {
	var result []string
	start := 0

	for i, char := range input {
		if i > 0 && unicode.IsUpper(char) {
			result = append(result, input[start:i])
			start = i
		}
	}

	// Add the last part of the string
	result = append(result, input[start:])

	return result
}

func convertToSnakeCase(input string) string {
	var result []rune

	for i, char := range input {
		if unicode.IsUpper(char) {
			// Add underscore if not at the beginning and the next character is not uppercase
			if i > 0 && i+1 < len(input) && !unicode.IsUpper(rune(input[i+1])) {
				result = append(result, '_')
			}
			// Convert the uppercase letter to lowercase
			result = append(result, unicode.ToLower(char))
		} else {
			result = append(result, char)
		}
	}

	return strings.ReplaceAll(string(result), ".", "_")
}

func convertToSnakeCase2(input string) string {
	var result []rune

	for i, char := range input {
		if i > 0 && char == '.' {
			// Replace subsequent dots with underscores
			result = append(result, '_')
		} else if unicode.IsUpper(char) {
			// Convert the uppercase letter to lowercase
			result = append(result, unicode.ToLower(char))
		} else {
			result = append(result, char)
		}
	}

	return string(result)
}

func convertToSnakeCase3(e any) string {
	input := fmt.Sprintf("%T", e)
	var result []rune

	sep := 'X'
	skip := true
	for i, char := range input {
		fmt.Println(i, string(char), skip)
		if skip && char != '.' {
			continue
		}
		if skip && char == '.' {
			skip = false
		}
		if unicode.IsUpper(char) {
			// Add underscore if not at the beginning and the next character is not uppercase
			if i > 0 && i+1 < len(input) && !unicode.IsUpper(rune(input[i+1])) {
				result = append(result, sep)
				sep = '_'
			}
			// Convert the uppercase letter to lowercase
			result = append(result, unicode.ToLower(char))
		} else {
			result = append(result, char)
		}
	}

	return string(result)
}

func main() {
	things := []any{
		NamespaceSomethingElseHappened{},
		NamespaceThingHappened{},
		events.NetSessionClosed{},
		events.NetCheckinHeard{},
	}

	for _, thing := range things {
		fmt.Println(convertToSnakeCase3(thing))
		return
	}
}
