package main

import (
	"strconv"
	"strings"
)

func getContrastColor(hexColor string) string {
	if strings.HasPrefix(hexColor, "#") {
		hexColor = hexColor[1:]
	}

	if len(hexColor) != 6 {
		return "#000000"
	}

	r, _ := strconv.ParseInt(hexColor[0:2], 16, 64)
	g, _ := strconv.ParseInt(hexColor[2:4], 16, 64)
	b, _ := strconv.ParseInt(hexColor[4:6], 16, 64)

	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255

	if luminance > 0.5 {
		return "#000000"
	}
	return "#FFFFFF"
}

func getPathColors(config Config, path []string) []string {
	colors := make([]string, len(path))

	if len(path) == 0 {
		return colors
	}

	var currentGroup *Group
	for _, group := range config.Groups {
		if group.Name == path[0] {
			currentGroup = group
			colors[0] = group.Color
			if colors[0] == "" {
				colors[0] = "#FFFFFF"
			}
			break
		}
	}

	if currentGroup == nil {
		return colors
	}

	for i := 1; i < len(path); i++ {
		found := false
		for _, subgroup := range currentGroup.Groups {
			if subgroup.Name == path[i] {
				currentGroup = subgroup
				colors[i] = subgroup.Color
				if colors[i] == "" {
					colors[i] = colors[i-1]
				}
				found = true
				break
			}
		}

		if !found {
			break
		}
	}

	return colors
}
