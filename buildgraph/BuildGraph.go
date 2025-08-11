package buildgraph

import (
	"github.com/julebarn/BSc-build-systems/buildinfo"
)

type BuildGraph struct {
	nodes map[string]*BuildGraphNode
}

func NewBuildGraph() *BuildGraph {
	return &BuildGraph{
		nodes: make(map[string]*BuildGraphNode),
	}
}

func (bg *BuildGraph) MakeNode(targetFilePath string, info buildinfo.Info) *BuildGraphNode {
	node := &BuildGraphNode{
		TargetFilePath: targetFilePath,
		Info:           info,
	}
	bg.nodes[targetFilePath] = node
	return node
}

type BuildGraphNode struct {
	TargetFilePath string
	FromCache      bool

	buildinfo.Info

	Dependencies []*BuildGraphNode
}

func (node *BuildGraphNode) AddDependency(dep *BuildGraphNode) {
	node.Dependencies = append(node.Dependencies, dep)
}

func (bg *BuildGraph) CalculateBuildOrder() []*BuildGraphNode {
	// This function uses topological sort to determine the build order of the nodes.
	// since this is a directed acyclic graph (DAG), we can use DFS-based approach.

	visited := make(map[string]bool)

	var orderedNodes []*BuildGraphNode
	for _, node := range bg.nodes {
		visit(node, &orderedNodes, visited)
	}
	return orderedNodes
}

func visit(node *BuildGraphNode, orderedNodes *[]*BuildGraphNode, visited map[string]bool) {
	if node.FromCache {
		return
	}

	if visited[node.TargetFilePath] {
		return
	}
	visited[node.TargetFilePath] = true

	for _, dep := range node.Dependencies {
		visit(dep, orderedNodes, visited)
	}
	*orderedNodes = append(*orderedNodes, node)
}
