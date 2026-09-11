package naming

import (
	"strings"
)

var irregularPlurals = map[string]string{
	"person":   "people",
	"man":      "men",
	"woman":    "women",
	"child":    "children",
	"datum":    "data",
	"status":   "statuses",
	"bus":      "buses",
	"quiz":     "quizzes",
	"matrix":   "matrices",
	"vertex":   "vertices",
	"index":    "indices",
	"criterion":"criteria",
	"analysis": "analyses",
	"thesis":   "theses",
}

var irregularSingulars = map[string]string{
	"people":   "person",
	"men":      "man",
	"women":    "woman",
	"children": "child",
	"data":     "datum",
	"statuses": "status",
	"buses":    "bus",
	"quizzes":  "quiz",
	"matrices": "matrix",
	"vertices": "vertex",
	"indices":  "index",
	"criteria": "criterion",
	"analyses": "analysis",
	"theses":   "thesis",
}

// Pluralize converts a singular English noun to its plural form.
func Pluralize(word string) string {
	if word == "" {
		return ""
	}

	lower := strings.ToLower(word)
	if plural, ok := irregularPlurals[lower]; ok {
		return matchCase(word, plural)
	}

	// -y preceded by a consonant -> -ies (e.g. category -> categories)
	if strings.HasSuffix(lower, "y") && len(lower) > 1 {
		prev := lower[len(lower)-2]
		if !isVowel(prev) {
			return word[:len(word)-1] + matchCaseSuffix(word, "ies")
		}
	}

	// -ch, -sh, -x, -z, -s, -ss -> -es (e.g. branch -> branches, pass -> passes)
	if strings.HasSuffix(lower, "ch") ||
		strings.HasSuffix(lower, "sh") ||
		strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z") ||
		strings.HasSuffix(lower, "s") {
		return word + matchCaseSuffix(word, "es")
	}

	// default rule: append -s
	return word + matchCaseSuffix(word, "s")
}

// Singularize converts a plural English noun to its singular form.
func Singularize(word string) string {
	if word == "" {
		return ""
	}

	lower := strings.ToLower(word)
	if singular, ok := irregularSingulars[lower]; ok {
		return matchCase(word, singular)
	}

	// -ies -> -y (e.g. categories -> category)
	if strings.HasSuffix(lower, "ies") && len(lower) > 3 {
		return word[:len(word)-3] + matchCaseSuffix(word, "y")
	}

	// -es suffixes: check for -ches, -shes, -xes, -zes, -ses
	if strings.HasSuffix(lower, "es") && len(lower) > 2 {
		stem := lower[:len(lower)-2]
		if strings.HasSuffix(stem, "ch") ||
			strings.HasSuffix(stem, "sh") ||
			strings.HasSuffix(stem, "x") ||
			strings.HasSuffix(stem, "z") ||
			strings.HasSuffix(stem, "ss") {
			return word[:len(word)-2]
		}
	}

	// -s (not -ss, -us, -is)
	if strings.HasSuffix(lower, "s") && !strings.HasSuffix(lower, "ss") && !strings.HasSuffix(lower, "us") && !strings.HasSuffix(lower, "is") {
		return word[:len(word)-1]
	}

	return word
}

func isVowel(b byte) bool {
	return b == 'a' || b == 'e' || b == 'i' || b == 'o' || b == 'u'
}

func matchCase(original, target string) string {
	if len(original) == 0 || len(target) == 0 {
		return target
	}
	if strings.ToUpper(original) == original {
		return strings.ToUpper(target)
	}
	if strings.ToUpper(original[:1]) == original[:1] {
		return strings.ToUpper(target[:1]) + strings.ToLower(target[1:])
	}
	return strings.ToLower(target)
}

func matchCaseSuffix(original, suffix string) string {
	if strings.ToUpper(original) == original {
		return strings.ToUpper(suffix)
	}
	return strings.ToLower(suffix)
}
