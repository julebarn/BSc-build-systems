package dependencybuilder

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/julebarn/BSc-build-systems/buildinfo"
	"github.com/julebarn/BSc-build-systems/dependencygraph"
)

func ReadJSONDependencyGraph(path string) *dependencygraph.DependencyGraphBuilder {

	var jsonGraph []DependencyGraphJSON

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	json.Unmarshal(data, &jsonGraph)

	fmt.Println(jsonGraph)

	var depGraph = dependencygraph.NewDependencyGraph()

	var nodeMap = make(map[string]*dependencygraph.DependencyGraphNode)
	var depList [][2]string

	for _, node := range jsonGraph {

		if node.IsSourceFile{
			fmt.Println("Adding source file: ", node.TargetFilePath)
		}

		depNode := depGraph.AddNode(
			node.TargetFilePath,
			buildinfo.Info{
				IsSourceFile:   node.IsSourceFile,
				DockerImage:    node.DockerImage,
				BuildCommand:   node.BuildCommand,
				OutputFilePath: node.OutputFilePath,
			},
		)

		nodeMap[node.TargetFilePath] = depNode

		for _, dep := range node.Dependencies {
			depList = append(depList, [2]string{node.TargetFilePath, dep})
		}
	}

	for _, dep := range depList {
		if depNode, exists := nodeMap[dep[1]]; exists {
			nodeMap[dep[0]].Dependencies = append(nodeMap[dep[0]].Dependencies, depNode)
		}
	}
	
	return depGraph

}

type DependencyGraphJSON struct {
	TargetFilePath string   `json:"target_file_path,omitempty"`
	Dependencies   []string `json:"dependencies,omitempty"`

	IsSourceFile bool `json:"is_source_file,omitempty"`

	DockerImage    string `json:"docker_image,omitempty"`
	BuildCommand   string `json:"build_command,omitempty"`
	OutputFilePath string `json:"output_file_path,omitempty"`
}
