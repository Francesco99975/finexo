package tools

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/labstack/gommon/log"
)

func tickerExtractor(seed string) (string, string, error) {
	log.Debugf("Extracting ticker and exchange from seed: %s", seed)
	if len(seed) <= 0 {
		return "", "", fmt.Errorf("seed is empty")
	}

	seed = strings.ToUpper(seed)
	seed = strings.TrimSpace(seed)
	seed = strings.ReplaceAll(seed, "/", "-")

	if strings.Contains(seed, ":") {
		parts := strings.Split(seed, ":")
		if len(parts) == 2 {
			return parts[1], parts[0], nil
		}
	} else if strings.Contains(seed, ".") {
		parts := strings.Split(seed, ".")
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
	}
	return seed, "", nil

}

func isAnEmptyString(s string) bool {
	s = strings.TrimSpace(s)
	return s == "" || s == "N/A" || s == "-" || s == "--" || s == "n/a"
}

func operandByFrequency(freq *string) int {
	switch *freq {
	case "weekly":
		return 52
	case "monthly":
		return 12
	case "quarterly":
		return 4
	case "semi-annual":
		return 2
	case "annual":
		return 1
	default:
		return 0
	}
}

func extractPercentage(input string) string {
	// Regular expression to capture content inside parentheses
	re := regexp.MustCompile(`\(([^)]+)\)`)
	match := re.FindStringSubmatch(input)

	// If a match is found, return the captured group
	if len(match) > 1 {
		return match[1] // Extracted percentage (without parentheses)
	}
	return "" // Return empty string if no match found
}
