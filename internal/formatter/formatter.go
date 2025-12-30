package formatter

import (
	"strings"
)

func ConvertToHosts(content string) (string, error) {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)
		
		if line == "" {
			result = append(result, "")
			continue
		}

		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			result = append(result, originalLine)
			continue
		}

		if strings.HasPrefix(line, "@@") {
			continue
		}

		if strings.HasPrefix(line, "||") {
			domain := strings.TrimPrefix(line, "||")
			
			if strings.HasSuffix(domain, "^") {
				domain = strings.TrimSuffix(domain, "^")
			} else if strings.HasSuffix(domain, ".") {
				domain = strings.TrimSuffix(domain, ".")
			}
			
			if domain != "" {
				result = append(result, "0.0.0.0 "+domain)
			}
			continue
		}

		if strings.HasPrefix(line, "/") && strings.HasSuffix(line, "/") {
			continue
		}

		result = append(result, originalLine)
	}

	return strings.Join(result, "\n"), nil
}
