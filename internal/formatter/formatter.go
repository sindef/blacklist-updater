package formatter

import (
	"fmt"
	"net"
	"strings"
)

func ConvertToHosts(content string, whitelist []string) (string, error) {
	return convert(content, true, false, "hosts", whitelist)
}

func ConvertToDNSmasq(content string, whitelist []string) (string, error) {
	return convert(content, false, true, "dnsmasq", whitelist)
}

func ConvertToRFC1035(content string, whitelist []string) (string, error) {
	result, err := convert(content, true, false, "rfc1035", whitelist)
	if err != nil {
		return "", err
	}
	
	soaRecord := "@\tIN\tSOA\tlocalhost. root.localhost. (\n\t\t1\t\t; serial\n\t\t3600\t\t; refresh\n\t\t1800\t\t; retry\n\t\t604800\t\t; expire\n\t\t86400\t\t; minimum TTL\n)\n"
	
	return soaRecord + result, nil
}

func convert(content string, stripRegex bool, keepWildcard bool, format string, whitelist []string) (string, error) {
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
			if format == "rfc1035" {
				if strings.HasPrefix(line, "!") {
					result = append(result, strings.Replace(originalLine, "!", ";", 1))
				} else {
					result = append(result, strings.Replace(originalLine, "#", ";", 1))
				}
			} else {
				result = append(result, originalLine)
			}
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
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ip := parts[0]
				if net.ParseIP(ip) != nil || isValidIPv4(ip) {
					domain = strings.Join(parts[1:], " ")
				} else {
					domain = line
				}
			} else {
				domain = line
			}
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
				if isWhitelisted(domain, whitelist) {
					continue
				}
				
				switch format {
				case "dnsmasq":
					result = append(result, "address=/"+domain+"/0.0.0.0")
				case "rfc1035":
					if !strings.HasSuffix(domain, ".") {
						domain = domain + "."
					}
					result = append(result, domain+"\tIN\tA\t0.0.0.0")
				default:
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

func isValidIPv4(ip string) bool {
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

func FilterWhitelist(content string, whitelist []string) string {
	if len(whitelist) == 0 {
		return content
	}
	
	lines := strings.Split(content, "\n")
	var result []string
	
	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)
		
		if line == "" {
			result = append(result, originalLine)
			continue
		}
		
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			result = append(result, originalLine)
			continue
		}
		
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			ip := parts[0]
			if net.ParseIP(ip) != nil || isValidIPv4(ip) {
				domain := strings.Join(parts[1:], " ")
				if isWhitelisted(domain, whitelist) {
					continue
				}
			}
		}
		
		result = append(result, originalLine)
	}
	
	return strings.Join(result, "\n")
}

func isWhitelisted(domain string, whitelist []string) bool {
	if len(whitelist) == 0 {
		return false
	}
	
	domain = strings.TrimSuffix(domain, ".")
	
	for _, pattern := range whitelist {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		
		if strings.HasPrefix(pattern, "*.") {
			suffix := strings.TrimPrefix(pattern, "*.")
			if strings.HasSuffix(domain, "."+suffix) || domain == suffix {
				return true
			}
		} else if domain == pattern {
			return true
		}
	}
	
	return false
}
