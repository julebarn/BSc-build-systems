package dependencygraph

import (
	"fmt"
	"strings"

)

func (tree *DependencyGraph) ToGraphviz() string {
	var sb strings.Builder
	sb.WriteString("digraph G {\n")

	for _, node := range tree.Nodes {
		color := "black"
		if node.NeedsUpdate {
			color = "red"
		}
		if node.BuildInfo.IsSourceFile {
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