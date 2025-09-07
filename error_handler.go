package main

import (
	"fmt"
	"strings"
)

type ErrorInfo struct {
	Message string
	Emoji   string
	Color   string
}

func classifyError(errMsg string) ErrorInfo {
	errLower := strings.ToLower(errMsg)

	var errorInfo ErrorInfo

	displayMsg := extractDisplayMessage(errMsg)
	isUsingExpiredCache := strings.Contains(errMsg, "[EXPIRED_CACHE]")

	if !isUsingExpiredCache {
		errorInfo.Color = "#FF6B6B"
	} else {
		errorInfo.Color = "#FFA726"
	}

	switch {
	case strings.Contains(errLower, "web imports disabled"):
		errorInfo.Emoji = "🚫"
		errorInfo.Color = "#FF6B6B"
		errorInfo.Message = fmt.Sprintf("Web imports disabled: %s", displayMsg)
		
	case strings.Contains(errLower, "no such host") || strings.Contains(errLower, "host not found") ||
		strings.Contains(errLower, "dns") || strings.Contains(errLower, "name resolution"):
		errorInfo.Emoji = "🌐"
		errorInfo.Message = fmt.Sprintf("Host not found: %s", displayMsg)

	case strings.Contains(errLower, "timeout") || strings.Contains(errLower, "connection timed out") ||
		strings.Contains(errLower, "deadline exceeded") || strings.Contains(errLower, "context deadline"):
		errorInfo.Emoji = "⏱️"
		errorInfo.Message = fmt.Sprintf("Connection timeout: %s", displayMsg)

	case strings.Contains(errLower, "connection refused") || strings.Contains(errLower, "connection reset") ||
		strings.Contains(errLower, "connection closed") || strings.Contains(errLower, "broken pipe"):
		errorInfo.Emoji = "🚫"
		errorInfo.Message = fmt.Sprintf("Connection refused: %s", displayMsg)

	case strings.Contains(errLower, "certificate") || strings.Contains(errLower, "ssl") || strings.Contains(errLower, "tls") ||
		strings.Contains(errLower, "x509") || strings.Contains(errLower, "handshake"):
		errorInfo.Emoji = "🔒"
		errorInfo.Message = fmt.Sprintf("SSL/TLS error: %s", displayMsg)

	case strings.Contains(errLower, "unauthorized") || strings.Contains(errLower, "auth"):
		errorInfo.Emoji = "🔐"
		errorInfo.Message = fmt.Sprintf("Authentication error: %s", displayMsg)

	case strings.Contains(errLower, "http error") || strings.Contains(errLower, "status") ||
		strings.Contains(errLower, "404") || strings.Contains(errLower, "not found") ||
		strings.Contains(errLower, "403") || strings.Contains(errLower, "forbidden") ||
		strings.Contains(errLower, "500") || strings.Contains(errLower, "internal server error") ||
		strings.Contains(errLower, "502") || strings.Contains(errLower, "bad gateway") ||
		strings.Contains(errLower, "503") || strings.Contains(errLower, "service unavailable") ||
		strings.Contains(errLower, "504") || strings.Contains(errLower, "gateway timeout") ||
		strings.Contains(errLower, "400") || strings.Contains(errLower, "bad request") ||
		strings.Contains(errLower, "429") || strings.Contains(errLower, "too many requests") ||
		strings.Contains(errLower, "301") || strings.Contains(errLower, "302") || strings.Contains(errLower, "redirect"):

		var specificMsg string
		switch {
		case strings.Contains(errLower, "404") || strings.Contains(errLower, "not found"):
			errorInfo.Emoji = "🔍"
			specificMsg = "Page not found (404)"
		case strings.Contains(errLower, "403") || strings.Contains(errLower, "forbidden"):
			errorInfo.Emoji = "🚫"
			specificMsg = "Access forbidden (403)"
		case strings.Contains(errLower, "401") || strings.Contains(errLower, "unauthorized"):
			errorInfo.Emoji = "🔐"
			specificMsg = "Authentication required (401)"
		case strings.Contains(errLower, "500") || strings.Contains(errLower, "internal server error"):
			errorInfo.Emoji = "💥"
			specificMsg = "Server error (500)"
		case strings.Contains(errLower, "502") || strings.Contains(errLower, "bad gateway"):
			errorInfo.Emoji = "🚧"
			specificMsg = "Bad gateway (502)"
		case strings.Contains(errLower, "503") || strings.Contains(errLower, "service unavailable"):
			errorInfo.Emoji = "🔧"
			specificMsg = "Service unavailable (503)"
		case strings.Contains(errLower, "504") || strings.Contains(errLower, "gateway timeout"):
			errorInfo.Emoji = "⏰"
			specificMsg = "Gateway timeout (504)"
		case strings.Contains(errLower, "400") || strings.Contains(errLower, "bad request"):
			errorInfo.Emoji = "❌"
			specificMsg = "Bad request (400)"
		case strings.Contains(errLower, "429") || strings.Contains(errLower, "too many requests"):
			errorInfo.Emoji = "🚦"
			specificMsg = "Rate limited (429)"
		case strings.Contains(errLower, "301") || strings.Contains(errLower, "302") || strings.Contains(errLower, "redirect"):
			errorInfo.Emoji = "↗️"
			specificMsg = "Redirect error"
		default:
			errorInfo.Emoji = "🌐"
			specificMsg = "HTTP error"
		}
		errorInfo.Message = fmt.Sprintf("%s: %s", specificMsg, displayMsg)

	case strings.Contains(errLower, "no such file") || strings.Contains(errLower, "file not found") ||
		strings.Contains(errLower, "permission denied") || strings.Contains(errLower, "access denied"):
		var specificMsg string
		switch {
		case strings.Contains(errLower, "permission denied") || strings.Contains(errLower, "access denied"):
			errorInfo.Emoji = "🔒"
			specificMsg = "Permission denied"
		default:
			errorInfo.Emoji = "📄"
			specificMsg = "File not found"
		}
		errorInfo.Message = fmt.Sprintf("%s: %s", specificMsg, displayMsg)

	case strings.Contains(errLower, "yaml") || strings.Contains(errLower, "parse") || strings.Contains(errLower, "unmarshal") ||
		strings.Contains(errLower, "invalid") || strings.Contains(errLower, "malformed"):
		if !strings.Contains(errLower, "http") && !strings.Contains(errLower, "host") && !strings.Contains(errLower, "connection") {
			var specificMsg string
			switch {
			case strings.Contains(errLower, "invalid") || strings.Contains(errLower, "malformed"):
				errorInfo.Emoji = "⚠️"
				specificMsg = "Invalid format"
			default:
				errorInfo.Emoji = "📝"
				specificMsg = "YAML parsing error"
			}
			errorInfo.Message = fmt.Sprintf("%s: %s", specificMsg, displayMsg)
		} else {
			errorInfo.Emoji = "🌐"
			errorInfo.Message = fmt.Sprintf("Network error: %s", displayMsg)
		}

	default:
		errorInfo.Emoji = "◉"
		errorInfo.Message = fmt.Sprintf("Error: %s", displayMsg)
	}

	if isUsingExpiredCache {
		errorInfo.Message += " (using expired cache)"
	}

	return errorInfo
}

func extractDisplayMessage(errMsg string) string {
	if strings.Contains(errMsg, "http://") || strings.Contains(errMsg, "https://") {
		parts := strings.Fields(errMsg)
		for _, part := range parts {
			if strings.HasPrefix(part, "http") {
				return strings.TrimSuffix(part, ":")
			}
		}
	}

	if strings.Contains(errMsg, ".yaml") || strings.Contains(errMsg, ".yml") {
		parts := strings.Fields(errMsg)
		for _, part := range parts {
			if strings.Contains(part, ".ya") {
				return strings.TrimSuffix(part, ":")
			}
		}
	}

	cleanMsg := errMsg

	prefixes := []string{
		"Import error for ",
		"Failed to read import file ",
		"Failed to parse import file ",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(cleanMsg, prefix) {
			cleanMsg = strings.TrimPrefix(cleanMsg, prefix)
			break
		}
	}

	if strings.Contains(cleanMsg, ":") {
		parts := strings.Split(cleanMsg, ":")
		for i := len(parts) - 1; i >= 0; i-- {
			trimmed := strings.TrimSpace(parts[i])
			if trimmed != "" && len(trimmed) > 2 {
				cleanMsg = trimmed
				break
			}
		}
	}

	cleanMsg = strings.TrimSpace(cleanMsg)
	cleanMsg = strings.TrimSuffix(cleanMsg, ":")
	cleanMsg = strings.TrimSuffix(cleanMsg, ";")

	if len(cleanMsg) < 3 {
		cleanMsg = strings.TrimSpace(errMsg)
		cleanMsg = strings.TrimSuffix(cleanMsg, ":")
	}

	return cleanMsg
}
