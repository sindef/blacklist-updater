package formatter

import (
	"strings"
)

func ConvertToHosts(content string) (string, error) {
	return convert(content, true, false)
}

func ConvertToDNSmasq(content string) (string, error) {
	return convert(content, false, true)
}

func convert(content string, stripRegex bool, keepWildcard bool) (string, error) {
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

		if stripRegex && (strings.HasPrefix(line, "/") && strings.HasSuffix(line, "/")) {
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
			if keepWildcard {
				domain = line
			} else {
				domain = strings.TrimPrefix(line, "*")
			}
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
			
			if !keepWildcard && strings.Contains(domain, "*") {
				domain = strings.ReplaceAll(domain, "*", "")
				if !isValidDomain(domain) {
					continue
				}
			}
			
			if domain != "" && !strings.HasPrefix(domain, "/") {
				if keepWildcard {
					result = append(result, "address=/"+domain+"/0.0.0.0")
				} else {
					result = append(result, "0.0.0.0 "+domain)
				}
				continue
			}
		}

		if !stripRegex && strings.HasPrefix(line, "/") && strings.HasSuffix(line, "/") {
			continue
		}

		result = append(result, originalLine)
	}

	return strings.Join(result, "\n"), nil
}

func isValidDomain(domain string) bool {
	if domain == "" {
		return false
	}
	
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return false
	}
	
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}
	
	if strings.Contains(domain, "..") {
		return false
	}
	
	parts := strings.Split(domain, ".")
	hasValidLabel := false
	for _, part := range parts {
		if part == "" {
			return false
		}
		if strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return false
		}
		hasValidLabel = true
	}
	
	return hasValidLabel
}
