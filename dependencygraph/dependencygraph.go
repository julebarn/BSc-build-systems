package dependencygraph

import (
	"fmt"

	"github.com/julebarn/BSc-build-systems/buildgraph"
)

type DependencyGraph struct {
	Nodes       map[string]*DependencyGraphNode `json:"nodes,omitempty"`
	SourceFiles []*DependencyGraphNode          `json:"source_files,omitempty"`
}

func (tree *DependencyGraph) BuildGraphForTarget(targetFilePath string) (*buildgraph.BuildGraph, error) {

	node, exists := tree.Nodes[targetFilePath]
	if !exists {
		return nil, fmt.Errorf("target file %s not found in dependency graph", targetFilePath)
	}

	buildgraph := buildgraph.NewBuildGraph()



	buildDependent := buildgraph.MakeNode(node.TargetFilePath, node.BuildInfo)
	if err := DependencyToBuildGraphNode(node, buildDependent, buildgraph); err != nil {
		return nil, fmt.Errorf("failed to convert dependency graph node to build graph node: %w", err)
	}

	return buildgraph, nil
}

func DependencyToBuildGraphNode(dep *DependencyGraphNode, buildDependent *buildgraph.BuildGraphNode, buildGraph *buildgraph.BuildGraph) error {
	
	if !dep.NeedsUpdate {
		buildDependent.FromCache = true
		buildDependent.Info = dep.BuildInfo
		buildDependent.TargetFilePath = dep.TargetFilePath
		return nil
	} else {
		buildDependent.FromCache = false
	}

	if dep.TargetFilePath != buildDependent.TargetFilePath {
		return fmt.Errorf("dependency target file path %s does not match build dependent target file path %s", dep.TargetFilePath, buildDependent.TargetFilePath)
	}

	for _, depNode := range dep.Dependencies {
	
		buildDepNode := buildGraph.MakeNode(depNode.TargetFilePath, depNode.BuildInfo)
	
		err := DependencyToBuildGraphNode(depNode, buildDepNode, buildGraph)
		if err != nil {
			return fmt.Errorf("failed to convert dependency %s to build graph node: %w", depNode.TargetFilePath, err)
		}

		buildDependent.AddDependency(buildDepNode)
	}

	return nil
}
