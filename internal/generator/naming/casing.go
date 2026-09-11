package naming

import (
	"strings"
	"unicode"
)

// SplitWords splits an identifier or phrase into lowercase words.
// Handles snake_case, kebab-case, camelCase, PascalCase, and spaces.
func SplitWords(s string) []string {
	var words []string
	var current strings.Builder

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '_' || r == '-' || r == ' ' || r == '/' || r == '.' {
			if current.Len() > 0 {
				words = append(words, strings.ToLower(current.String()))
				current.Reset()
			}
			continue
		}

		if unicode.IsUpper(r) {
			// Check if new word boundary:
			// 1. Preceding character was lower/digit (e.g. "myVar")
			// 2. Or part of an acronym ending and next char is lower (e.g. "HTTPClient" -> "HTTP", "Client")
			if current.Len() > 0 {
				prevIsUpper := i > 0 && unicode.IsUpper(runes[i-1])
				nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])

				if !prevIsUpper || (prevIsUpper && nextIsLower) {
					words = append(words, strings.ToLower(current.String()))
					current.Reset()
				}
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, strings.ToLower(current.String()))
	}

	return words
}

// ToPascalCase converts an identifier to PascalCase (e.g. "user_account" -> "UserAccount").
func ToPascalCase(s string) string {
	words := SplitWords(s)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	for _, w := range words {
		if len(w) == 0 {
			continue
		}
		b.WriteString(strings.ToUpper(w[:1]))
		b.WriteString(w[1:])
	}
	return b.String()
}

// ToCamelCase converts an identifier to camelCase (e.g. "user_account" -> "userAccount").
func ToCamelCase(s string) string {
	words := SplitWords(s)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		if i == 0 {
			b.WriteString(strings.ToLower(w))
		} else {
			b.WriteString(strings.ToUpper(w[:1]))
			b.WriteString(w[1:])
		}
	}
	return b.String()
}

// ToSnakeCase converts an identifier to snake_case (e.g. "UserAccount" -> "user_account").
func ToSnakeCase(s string) string {
	words := SplitWords(s)
	return strings.Join(words, "_")
}

// ToKebabCase converts an identifier to kebab-case (e.g. "UserAccount" -> "user-account").
func ToKebabCase(s string) string {
	words := SplitWords(s)
	return strings.Join(words, "-")
}

// ToPackageName converts an identifier to a canonical Go package name (e.g. "UserAccounts" -> "useraccounts").
func ToPackageName(s string) string {
	words := SplitWords(s)
	return strings.Join(words, "")
}
