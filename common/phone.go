package common

import "strings"

func NormalizePhoneNumber(raw string, defaultCountryCode string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	var digits strings.Builder
	hasLeadingPlus := false
	for i, r := range raw {
		switch {
		case r >= '0' && r <= '9':
			digits.WriteRune(r)
		case r == '+' && i == 0:
			hasLeadingPlus = true
		}
	}

	normalizedDigits := digits.String()
	if normalizedDigits == "" {
		return ""
	}
	if strings.HasPrefix(normalizedDigits, "00") {
		hasLeadingPlus = true
		normalizedDigits = strings.TrimPrefix(normalizedDigits, "00")
	}
	if hasLeadingPlus {
		if len(normalizedDigits) < 6 || len(normalizedDigits) > 18 {
			return ""
		}
		return "+" + normalizedDigits
	}

	cc := normalizeCountryCode(defaultCountryCode)
	if cc == "" {
		if len(normalizedDigits) < 6 || len(normalizedDigits) > 18 {
			return ""
		}
		return "+" + normalizedDigits
	}
	normalizedDigits = strings.TrimPrefix(normalizedDigits, "0")
	if normalizedDigits == "" {
		return ""
	}
	if len(cc)+len(normalizedDigits) < 6 || len(cc)+len(normalizedDigits) > 18 {
		return ""
	}
	return "+" + cc + normalizedDigits
}

func normalizeCountryCode(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var digits strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	return digits.String()
}

func LooksLikeEmailIdentifier(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	if strings.Contains(raw, "@") {
		return true
	}
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}
