package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DefaultProjectsDir is the default directory for all projects
	DefaultProjectsDir = "PROJECTS"
)

// OutputConfig holds configuration for file output
type OutputConfig struct {
	ProjectName string
	OutputFile  string
	BaseDir     string
}

// NewOutputConfig creates a new output configuration
func NewOutputConfig(projectName, outputFile string) *OutputConfig {
	return &OutputConfig{
		ProjectName: projectName,
		OutputFile:  outputFile,
		BaseDir:     DefaultProjectsDir,
	}
}

// GetOutputPath returns the full path for output file
// If project is specified, it creates: PROJECTS/<project>/<filename>
// If only output is specified, it creates: PROJECTS/<filename>
// If neither is specified, it returns an empty string
func (c *OutputConfig) GetOutputPath(defaultFilename string) (string, error) {
	if c.ProjectName == "" && c.OutputFile == "" {
		return "", nil
	}

	// Start with base directory
	basePath := c.BaseDir

	// Add project subdirectory if specified
	if c.ProjectName != "" {
		basePath = filepath.Join(basePath, c.ProjectName)
		if os.Getenv("GENGO_DEBUG") != "" {
			fmt.Printf("DEBUG GetOutputPath: Adding project dir, basePath=%q\n", basePath)
		}
	}

	// Create the directory structure
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %v", basePath, err)
	}

	// Determine filename
	filename := c.OutputFile
	if filename == "" {
		filename = defaultFilename
	}

	// Return full path
	return filepath.Join(basePath, filename), nil
}

// SaveToProject saves content to the project structure
func (c *OutputConfig) SaveToProject(content []byte, defaultFilename string) (string, error) {
	outputPath, err := c.GetOutputPath(defaultFilename)
	if err != nil {
		return "", err
	}

	if outputPath == "" {
		return "", fmt.Errorf("no output path specified")
	}

	// Debug: Print what we're doing
	if os.Getenv("GENGO_DEBUG") != "" {
		fmt.Printf("DEBUG SaveToProject: ProjectName=%q, OutputFile=%q, DefaultFile=%q, Path=%q\n", 
			c.ProjectName, c.OutputFile, defaultFilename, outputPath)
	}

	// Write content to file
	if err := os.WriteFile(outputPath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}

	return outputPath, nil
}

// GetProjectDir returns the project directory path
func (c *OutputConfig) GetProjectDir() string {
	if c.ProjectName != "" {
		return filepath.Join(c.BaseDir, c.ProjectName)
	}
	return c.BaseDir
}