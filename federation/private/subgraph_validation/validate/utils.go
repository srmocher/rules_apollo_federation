package main

import (
	"archive/tar"
	"fmt"
	"github.com/bazelbuild/rules_go/go/runfiles"
	"io"
	"os"
	"path/filepath"
	"runtime"

	yaml "gopkg.in/yaml.v3"
)

type apolloSupergraphConfig struct {
	FederationVersion string              `yaml:"federation_version"`
	Subgraphs         map[string]subgraph `yaml:"subgraphs"`
}

type subgraph struct {
	RoutingUrl string         `yaml:"routing_url"`
	Schema     subgraphSchema `yaml:"schema"`
}

type subgraphSchema struct {
	File string `yaml:"file"`
}

func extractTar(tarFile string, destDir string) error {
	// Open the tar file
	file, err := os.Open(tarFile)
	if err != nil {
		return fmt.Errorf("failed to open tar file: %w", err)
	}
	defer file.Close()

	// Create a tar reader
	tr := tar.NewReader(file)

	// Iterate through the files in the tar archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of tar archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar file: %w", err)
		}

		// Determine the type of the file
		targetPath := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create the directory
			err = os.MkdirAll(targetPath, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			// Create the file
			err = writeFile(targetPath, tr, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", targetPath, err)
			}
		default:
			// Skip other types (e.g., symlinks)
			fmt.Printf("Skipping unsupported type: %s\n", header.Name)
		}
	}

	return nil
}

func writeFile(filePath string, reader io.Reader, mode os.FileMode) error {
	// Create the parent directories if necessary
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create parent directories for %s: %w", filePath, err)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	// Set file permissions
	if err := file.Chmod(mode); err != nil {
		return fmt.Errorf("failed to set permissions for %s: %w", filePath, err)
	}

	// Write the file contents
	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("failed to write file data for %s: %w", filePath, err)
	}

	return nil
}

func roverCliPath() (string, error) {
	r, err := runfiles.New()
	if err != nil {
		return "", err
	}

	platformSuffix := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	return r.Rlocation(fmt.Sprintf("apollo_rover_%s/rover", platformSuffix))
}

func parseRoverConfigFile(configFile string) (*apolloSupergraphConfig, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	var asc apolloSupergraphConfig

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&asc); err != nil {
		return nil, err
	}

	return &asc, nil
}
