package validator

import (
	"fmt"
	"net"
	"strings"
)

func ValidateHostsFile(content string) bool {
	if len(content) == 0 {
		return false
	}

	lines := strings.Split(content, "\n")
	hasValidEntry := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		if isValidAdGuardRule(line) {
			hasValidEntry = true
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		ip := parts[0]
		if !isValidIP(ip) {
			return false
		}

		hasValidEntry = true
	}

	return hasValidEntry
}

func isValidAdGuardRule(line string) bool {
	return strings.HasPrefix(line, "||") && strings.HasSuffix(line, "^")
}

func isValidIP(ip string) bool {
	if idx := strings.Index(ip, "%"); idx != -1 {
		ip = ip[:idx]
	}

	if net.ParseIP(ip) != nil {
		return true
	}

	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		var num int
		if _, err := fmt.Sscanf(part, "%d", &num); err != nil {
			return false
		}
		if num < 0 || num > 255 {
			return false
		}
	}
	return true
}
