package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/julebarn/BSc-build-systems/build"
	"github.com/julebarn/BSc-build-systems/buildinfo"
	"github.com/julebarn/BSc-build-systems/cache"
	"github.com/julebarn/BSc-build-systems/dependencybuilder"
)

func main() {

	target := readArgument()

	cacheDir := "./cache"
	c := cache.NewCache(cacheDir)

	bgBuilder := dependencybuilder.ReadJSONDependencyGraph("./build.json")

	DependencyGraph := bgBuilder.MakeDependencyGraph(c)

	fmt.Println("Build Graph:")
	fmt.Println(DependencyGraph.ToGraphviz())

	buildgraph, err := DependencyGraph.BuildGraphForTarget(target)

	if err != nil {
		fmt.Printf("Error building graph for target: %v\n", err)
		clearCache(cacheDir)

		return
	}
	fmt.Println("Build Graph for target:", target)
	fmt.Println(buildgraph.ToGraphviz())

	BuildOrder := buildgraph.CalculateBuildOrder()

	fmt.Println("Build Order:")
	for _, node := range BuildOrder {
		fmt.Printf("Node: %s, Build Info: %+v\n", node.TargetFilePath, node.Info)
	}

	b, err := build.NewBuildEnvironment(context.Background())
	if err != nil {
		fmt.Printf("Error creating build environment: %v\n", err)
		clearCache(cacheDir)
		return
	}

	for _, node := range BuildOrder {

		err := b.Build(node, c)
		if err != nil {
			fmt.Printf("Error building node %s: %v\n", node.TargetFilePath, err)
			clearCache(cacheDir)
			return
		}
	}


	fmt.Println("Build completed successfully!")
	out, hit, err := c.Get(target)
	if err != nil {
		fmt.Printf("Error getting output from cache: %v\n", err)
		clearCache(cacheDir)
		return
	}

	if !hit {
		fmt.Println("Cache miss for output file, writing to disk.")
	}

	err = os.WriteFile(out.TargetPath, out.File, 0644)
	if err != nil {
		clearCache(cacheDir)
		fmt.Printf("Error writing output file: %v\n", err)
		return
	}

}

func readArgument() string {
	if len(os.Args) < 2 {
		panic("no target specified")
	}
	return os.Args[1]
}


func SourceFile() buildinfo.Info {
	return buildinfo.Info{
		IsSourceFile: true,
	}
}

func gcc(cmd string) buildinfo.Info {
	return buildinfo.Info{
		IsSourceFile:   false,
		DockerImage:    "gcc:latest",
		BuildCommand:   "gcc " + cmd,
		OutputFilePath: "",
	}
}

func gccLink(outputFile string, inputFiles ...string) buildinfo.Info {
	return buildinfo.Info{
		IsSourceFile:   false,
		DockerImage:    "gcc:latest",
		BuildCommand:   fmt.Sprintf("gcc -o %s %s", outputFile, strings.Join(inputFiles, " ")),
		OutputFilePath: outputFile,
	}
}

func clearCache(cacheDir string) error {
	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("failed to clear cache directory %s: %w", cacheDir, err)
	}
	fmt.Println("Cache cleared.")
	return nil
}