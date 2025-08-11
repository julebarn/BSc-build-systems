package buildgraph

import (
	"fmt"
	"strings"

)

func (tree *BuildGraph) ToGraphviz() string {
	var sb strings.Builder
	sb.WriteString("digraph G {\n")

	for _, node := range tree.nodes {
		color := "black"
		if node.FromCache {
			color = "blue"
		}

		sb.WriteString(fmt.Sprintf("  %q [label=%q color=%q];\n", node.TargetFilePath, node.TargetFilePath, color))
		for _, dep := range node.Dependencies {
			sb.WriteString(fmt.Sprintf("  %q -> %q;\n", node.TargetFilePath, dep.TargetFilePath))
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}