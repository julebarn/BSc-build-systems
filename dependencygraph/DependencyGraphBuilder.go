package dependencygraph

import (
	"crypto/md5"
	"fmt"
	"os"

	"github.com/julebarn/BSc-build-systems/buildinfo"
	"github.com/julebarn/BSc-build-systems/cache"
)

type DependencyGraphBuilder struct {
	Nodes       map[string]*DependencyGraphNode
	SourceFiles []*DependencyGraphNode
}

func NewDependencyGraph() *DependencyGraphBuilder {
	return &DependencyGraphBuilder{
		Nodes: make(map[string]*DependencyGraphNode),
	}
}

func (tree *DependencyGraphBuilder) AddNode(targetFilePath string, buildInfo buildinfo.Info, dependencies ...*DependencyGraphNode) *DependencyGraphNode {
	node := &DependencyGraphNode{
		TargetFilePath: targetFilePath,
		BuildInfo:      buildInfo,
		Dependencies:   dependencies,
		Dependent:      []*DependencyGraphNode{},
		NeedsUpdate:    false,
	}

	if buildInfo.IsSourceFile {
		tree.SourceFiles = append(tree.SourceFiles, node)
	}

	tree.Nodes[targetFilePath] = node

	return node
}

func (tree *DependencyGraphBuilder) MakeDependencyGraph(filecache *cache.Cache) DependencyGraph {

	tree.calculateDependencies()

	tree.calculateNeedsUpdate(filecache)

	return DependencyGraph{
		Nodes:       tree.Nodes,
		SourceFiles: tree.SourceFiles,
	}

}

func (tree *DependencyGraphBuilder) calculateNeedsUpdate(filecache *cache.Cache) error {

	for _, node := range tree.SourceFiles {
		target, hit, err := filecache.Get(node.TargetFilePath)
		if err != nil {
			return fmt.Errorf("failed to get target from cache: %w", err)
		}

		if !hit {
			fmt.Println("Cache miss for:", node.TargetFilePath)
			node.needsUpdate(filecache)
			continue
		}

		fileHash, err := getTargetFileHash(node.TargetFilePath)
		if err != nil {
			return fmt.Errorf("failed to get file hash: %w", err)
		}

		if target.HashFile != fileHash {
			node.needsUpdate(filecache)
		}
	}

	return nil
}

func getTargetFileHash(path string) ([16]byte, error) {

	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading file:", path, err)
		return [16]byte{}, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return md5.Sum(file), nil
}

func (tree *DependencyGraphBuilder) calculateDependencies() {
	for _, node := range tree.Nodes {
		if len(node.Dependencies) == 0 {
			continue
		}

		for _, dep := range node.Dependencies {
			dep.Dependent = append(dep.Dependent, node)
		}
	}
}

type DependencyGraphNode struct {
	TargetFilePath string                
	BuildInfo      buildinfo.Info         
	Dependencies   []*DependencyGraphNode 

	Dependent   []*DependencyGraphNode 
	NeedsUpdate bool                   
}

func (node *DependencyGraphNode) needsUpdate(filecache *cache.Cache) {
	fmt.Println("Checking if node needs update:", node.TargetFilePath)
	// If the node already needs an update, no need to check further
	if node.NeedsUpdate {
		return
	}

	node.NeedsUpdate = true

	for _, dep := range node.Dependent {
		fmt.Println("Checking dependent node:", dep.TargetFilePath)
		dep.needsUpdate(filecache)
	}
}
