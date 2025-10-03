package tty

import (
	"regexp"
	"strings"
)

type PromptAction int

const (
	ActionSendPassword PromptAction = iota
	ActionIgnore
)

type PromptMatcher struct {
	patterns []PromptPattern
}

type PromptPattern struct {
	name     string
	patterns []string
	regexes  []*regexp.Regexp
	action   PromptAction
}

func NewPromptMatcher() *PromptMatcher {
	passwordPatterns := []string{
		"assword:",
		"Password:",
		"Enter password:",
		"login password:",
		"password for",
	}

	patterns := []PromptPattern{
		{
			name:     "password",
			patterns: passwordPatterns,
			action:   ActionSendPassword,
		},
	}

	for i := range patterns {
		for _, pattern := range patterns[i].patterns {
			if regex, err := regexp.Compile("(?i)" + regexp.QuoteMeta(pattern)); err == nil {
				patterns[i].regexes = append(patterns[i].regexes, regex)
			}
		}
	}

	return &PromptMatcher{patterns: patterns}
}

func (pm *PromptMatcher) MatchPrompt(buffer []byte) (PromptAction, bool) {
	text := string(buffer)

	for _, pattern := range pm.patterns {
		for _, str := range pattern.patterns {
			if strings.Contains(strings.ToLower(text), strings.ToLower(str)) {
				return pattern.action, true
			}
		}

		for _, regex := range pattern.regexes {
			if regex.MatchString(text) {
				return pattern.action, true
			}
		}
	}

	return ActionIgnore, false
}
