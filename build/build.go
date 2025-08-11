package build

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/julebarn/BSc-build-systems/buildgraph"
	"github.com/julebarn/BSc-build-systems/cache"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
)

type BuildEnvironment struct {
	dockerClient *client.Client

	ctx context.Context
}

func NewBuildEnvironment(ctx context.Context) (*BuildEnvironment, error) {
	// TODO/NB:  useing client.FromEnv to get Docker client configuration from environment variables.
	// is probably not the best way to garantie that the file build is correctly configured.

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &BuildEnvironment{
		dockerClient: dockerClient,

		ctx: ctx,
	}, nil
}

func (env *BuildEnvironment) Build(build *buildgraph.BuildGraphNode, c *cache.Cache) error {

	if build.IsSourceFile {
		return readAndCacheSourceFile(build, c)
	}
	return executeBuildProcess(env, build, c)
}

func executeBuildProcess(env *BuildEnvironment, build *buildgraph.BuildGraphNode, c *cache.Cache) error {

	err := env.pullImageifNeeded(env.ctx, build.Info.DockerImage)
	if err != nil {
		return fmt.Errorf("failed to pull Docker image %s: %w", build.Info.DockerImage, err)
	}

	resp, clean, err := createBuildContainer(build, env, &container.Config{
		Image: build.Info.DockerImage,
		Cmd:   strings.Fields(build.Info.BuildCommand),
	})
	if err != nil {
		return fmt.Errorf("failed to create container for build %s: %w", build.TargetFilePath, err)
	}
	defer clean()

	// Copy dependency files into the container
	err = copyDependenciesToContainer(build, c, env, resp)
	if err != nil {
		return err
	}

	err = env.dockerClient.ContainerStart(env.ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container for build %s: %w", build.TargetFilePath, err)
	}

	// TODO/NB: this is should not be a blocking call, but in stead run multiple builds in parallel.
	err = waitForContainer(env, resp)
	if err != nil {
		return err
	}

	fmt.Printf("Build completed for %s\n", build.TargetFilePath)
	//copy the output file from the container to the cache
	outputFile := build.Info.OutputFilePath
	if outputFile == "" {
		outputFile = build.TargetFilePath // Default to target file path if no output specified
	}

	outputReader, _, err := env.dockerClient.CopyFromContainer(env.ctx, resp.ID, outputFile)

	if err != nil {
		return fmt.Errorf("failed to copy output file %s from container: %w", outputFile, err)
	}
	defer outputReader.Close()

	tr := tar.NewReader(outputReader)
	var fileData bytes.Buffer

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar header: %w", err)
		}

		if header.Typeflag == tar.TypeReg {
			if _, err := io.Copy(&fileData, tr); err != nil {
				return fmt.Errorf("error extracting tar file: %w", err)
			}
			break
		}
	}

	cacheEntry := cache.NewTarget(outputFile, fileData.Bytes())

	err = c.Set(build.TargetFilePath, cacheEntry)
	if err != nil {
		return fmt.Errorf("failed to put output file %s into cache: %w", build.TargetFilePath, err)
	}

	return nil
}

func createBuildContainer(build *buildgraph.BuildGraphNode, env *BuildEnvironment, containerConfig *container.Config) (container.CreateResponse, func(), error) {
	fmt.Printf("Building %s with command: %s\n", build.TargetFilePath, build.Info.BuildCommand)

	resp, err := env.dockerClient.ContainerCreate(env.ctx, containerConfig, nil, nil, nil, "")
	if err != nil {
		return container.CreateResponse{}, nil, fmt.Errorf("failed to create container for build %s: %w", build.TargetFilePath, err)
	}

	clean := func() {
		if err := env.dockerClient.ContainerRemove(env.ctx, resp.ID, container.RemoveOptions{
			Force: true,
		}); err != nil {
			fmt.Printf("Failed to remove container %s: %v\n", resp.ID, err)
		}
	}
	return resp, clean, nil
}

func waitForContainer(env *BuildEnvironment, resp container.CreateResponse) error {
	statusCh, errCh := env.dockerClient.ContainerWait(env.ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return fmt.Errorf("container wait failed: %w", err)
	case <-statusCh:
		return nil
	}
}

func copyDependenciesToContainer(build *buildgraph.BuildGraphNode, c *cache.Cache, env *BuildEnvironment, resp container.CreateResponse) error {
	for _, dep := range build.Dependencies {
		FileCacheEntry, hit, err := c.Get(dep.TargetFilePath)
		if err != nil {
			return fmt.Errorf("failed to get cache entry for %s: %w", dep.TargetFilePath, err)
		}
		if !hit {
			fmt.Println("Cache miss for dependency:", dep.TargetFilePath)
		}

		fmt.Println("Copying dependency to container:", dep.TargetFilePath)

		err = env.dockerClient.CopyToContainer(
			env.ctx,
			resp.ID,
			"/",
			getTarFromCacheEntry(FileCacheEntry),
			container.CopyToContainerOptions{
				AllowOverwriteDirWithFile: true,
			})
		if err != nil {
			return fmt.Errorf("failed to copy dependency %s to container: %w", dep.TargetFilePath, err)
		}
	}
	return nil
}

func readAndCacheSourceFile(build *buildgraph.BuildGraphNode, c *cache.Cache) error {
	data, err := os.ReadFile(build.TargetFilePath)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", build.TargetFilePath, err)
	}

	cacheEntry := cache.NewTarget(build.TargetFilePath, data)
	err = c.Set(build.TargetFilePath, cacheEntry)
	if err != nil {
		return fmt.Errorf("failed to cache source file %s: %w", build.TargetFilePath, err)
	}

	return nil
}

func getTarFromCacheEntry(FileCacheEntry cache.FileCacheEntry) io.Reader {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	header := &tar.Header{
		Name:     FileCacheEntry.TargetPath,
		Size:     int64(len(FileCacheEntry.File)),
		Typeflag: tar.TypeReg,
	}

	if err := tw.WriteHeader(header); err != nil {
		fmt.Printf("Error writing tar header: %v\n", err)
		return nil
	}

	if _, err := tw.Write(FileCacheEntry.File); err != nil {
		fmt.Printf("Error writing data to tar: %v\n", err)
		return nil
	}

	if err := tw.Close(); err != nil {
		fmt.Printf("Error closing tar writer: %v\n", err)
		return nil
	}

	return &buf

}

func (env *BuildEnvironment) pullImageifNeeded(ctx context.Context, imageName string) error {
	imgLst, err := env.dockerClient.ImageList(ctx, image.ListOptions{
		All: true,
	})
	if err != nil {
		return err
	}

	for _, img := range imgLst {
		if img.RepoTags != nil && len(img.RepoTags) > 0 && img.RepoTags[0] == imageName {
			fmt.Printf("Image %s already exists, skipping pull.\n", imageName)
			return nil
		}
	}

	// If we reach here, the image is not found locally, so we need to pull it.
	fmt.Printf("Pulling Docker image: %s\n", imageName)
	reader, err := env.dockerClient.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader) // Show image pull progress

	return nil
}
