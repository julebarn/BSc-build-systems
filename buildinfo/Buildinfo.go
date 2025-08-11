package buildinfo


type Info struct {
	IsSourceFile bool `json:"is_source_file,omitempty"`

	DockerImage    string `json:"docker_image,omitempty"`
	BuildCommand   string `json:"build_command,omitempty"`
	OutputFilePath string `json:"output_file_path,omitempty"`
}
