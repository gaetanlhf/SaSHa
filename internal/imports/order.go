package imports

import (
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

func NewOrderTracker() *OrderTracker {
	return &OrderTracker{
		positions: make(map[string][]ElementPosition),
	}
}

func (ot *OrderTracker) ParseYAML(data []byte) error {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return err
	}

	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		rootNode := node.Content[0]
		ot.extractPositions(rootNode, "")
	}

	return nil
}

func (ot *OrderTracker) extractPositions(node *yaml.Node, parentPath string) {
	if node.Kind != yaml.MappingNode {
		return
	}

	var positions []ElementPosition
	importIdx := 0
	groupIdx := 0
	hostIdx := 0

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		if keyNode.Value == "inventory" && valueNode.Kind == yaml.MappingNode {
			ot.extractPositions(valueNode, parentPath)
			continue
		}

		switch keyNode.Value {
		case "imports":
			if valueNode.Kind == yaml.SequenceNode {
				for _, importNode := range valueNode.Content {
					positions = append(positions, ElementPosition{
						Type:  "import",
						Index: importIdx,
						Line:  importNode.Line,
					})
					importIdx++
				}
			}

		case "groups":
			if valueNode.Kind == yaml.SequenceNode {
				for _, groupNode := range valueNode.Content {
					positions = append(positions, ElementPosition{
						Type:  "group",
						Index: groupIdx,
						Line:  groupNode.Line,
					})

					groupName := extractNameFromNode(groupNode)
					if groupName != "" {
						newPath := parentPath
						if newPath != "" {
							newPath += "/"
						}
						newPath += groupName
						ot.extractPositions(groupNode, newPath)
					}

					groupIdx++
				}
			}

		case "hosts":
			if valueNode.Kind == yaml.SequenceNode {
				for _, hostNode := range valueNode.Content {
					positions = append(positions, ElementPosition{
						Type:  "host",
						Index: hostIdx,
						Line:  hostNode.Line,
					})
					hostIdx++
				}
			}
		}
	}

	sort.Slice(positions, func(i, j int) bool {
		return positions[i].Line < positions[j].Line
	})

	ot.positions[parentPath] = positions
}
func extractNameFromNode(node *yaml.Node) string {
	if node.Kind != yaml.MappingNode {
		return ""
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == "name" && i+1 < len(node.Content) {
			return node.Content[i+1].Value
		}
	}
	return ""
}

func (ot *OrderTracker) GetOrder(path string) []ElementPosition {
	return ot.positions[path]
}

func LoadOrderTracker(configPath string) *OrderTracker {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	tracker := NewOrderTracker()
	if err := tracker.ParseYAML(data); err != nil {
		return nil
	}

	return tracker
}

func mergeWithYAMLOrder[T any](existing []T, imported []T, path string, tracker *OrderTracker) []T {
	if tracker == nil {
		return append(existing, imported...)
	}

	order := tracker.GetOrder(path)
	if len(order) == 0 {
		return append(existing, imported...)
	}

	result := make([]T, 0, len(existing)+len(imported))
	existingIdx := 0
	importedIdx := 0

	for _, pos := range order {
		switch pos.Type {
		case "import":
			if importedIdx < len(imported) {
				result = append(result, imported[importedIdx])
				importedIdx++
			}
		case "group", "host":
			if existingIdx < len(existing) {
				result = append(result, existing[existingIdx])
				existingIdx++
			}
		}
	}

	result = append(result, existing[existingIdx:]...)
	result = append(result, imported[importedIdx:]...)

	return result
}
