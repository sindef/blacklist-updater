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

		if strings.HasPrefix(line, "-") {
			continue
		}

		var domain string
		if strings.HasPrefix(line, "||") {
			domain = strings.TrimPrefix(line, "||")
		} else if strings.HasPrefix(line, "|") {
			domain = strings.TrimPrefix(line, "|")
		} else if strings.HasPrefix(line, ".") {
			domain = strings.TrimPrefix(line, ".")
		} else if strings.HasPrefix(line, "://") {
			domain = strings.TrimPrefix(line, "://")
		} else if strings.HasPrefix(line, "^") {
			domain = strings.TrimPrefix(line, "^")
		} else if strings.HasPrefix(line, "*") {
			domain = strings.TrimPrefix(line, "*")
		} else {
			domain = line
		}

		if domain != "" {
			domain = strings.TrimSuffix(domain, "^|")
			domain = strings.TrimSuffix(domain, "^")
			domain = strings.TrimSuffix(domain, "|")
			domain = strings.TrimSuffix(domain, ".")
			
			if strings.Contains(domain, "://") {
				parts := strings.Split(domain, "://")
				if len(parts) > 1 {
					urlPart := parts[1]
					if idx := strings.Index(urlPart, "/"); idx != -1 {
						domain = urlPart[:idx]
					} else if idx := strings.Index(urlPart, "^"); idx != -1 {
						domain = urlPart[:idx]
					} else {
						domain = urlPart
					}
				}
			}
			
			if idx := strings.Index(domain, "/"); idx != -1 {
				domain = domain[:idx]
			}
			if idx := strings.Index(domain, "^"); idx != -1 {
				domain = domain[:idx]
			}
			
			if domain != "" && !strings.HasPrefix(domain, "/") {
				result = append(result, "0.0.0.0 "+domain)
				continue
			}
		}

		if strings.HasPrefix(line, "/") && strings.HasSuffix(line, "/") {
			continue
		}

		result = append(result, originalLine)
	}

	return strings.Join(result, "\n"), nil
}
