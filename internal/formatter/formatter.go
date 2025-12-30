package formatter

import (
	"strings"
)

func ConvertToHosts(content string) (string, error) {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if line == "" {
			result = append(result, "")
			continue
		}

		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			result = append(result, line)
			continue
		}

		if strings.HasPrefix(line, "||") && strings.HasSuffix(line, "^") {
			domain := strings.TrimPrefix(line, "||")
			domain = strings.TrimSuffix(domain, "^")
			if domain != "" {
				result = append(result, "0.0.0.0 "+domain)
			}
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n"), nil
}
