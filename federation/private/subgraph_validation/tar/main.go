package main

import (
	"archive/tar"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type subgraphSchemaPaths []string

func (s *subgraphSchemaPaths) String() string {
	return strings.Join(*s, ",")
}

func (s *subgraphSchemaPaths) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func addDirectoryToTar(tw *tar.Writer, dirPath string) error {
	header := &tar.Header{
		Name:    dirPath + string(filepath.Separator), // Add separator to indicate it's a directory
		Mode:    0755,                                 // Use directory permissions
		ModTime: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		Uid:     0,       // Use a fixed user ID
		Gid:     0,       // Use a fixed group ID
		Uname:   "bazel", // Use a fixed user name
		Gname:   "bazel", // Use a fixed group name
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write directory header for %s: %w", dirPath, err)
	}

	return nil
}

func addFiletoTar(tw *tar.Writer, filePath string, useFullPath bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	path := filePath
	if !useFullPath {
		path = filepath.Base(path)
	}
	header := &tar.Header{
		Name:    path,
		Size:    info.Size(),
		Mode:    0644, // Use fixed permissions
		ModTime: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		Uid:     0,       // Use a fixed user ID
		Gid:     0,       // Use a fixed group ID
		Uname:   "bazel", // Use a fixed user name
		Gname:   "bazel", // Use a fixed group name
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := io.Copy(tw, file); err != nil {
		return fmt.Errorf("failed to write file data: %w", err)
	}

	return nil
}

func createSubgraphsTarFile(tarFile string, paths subgraphSchemaPaths, roverConfigFile string, schemasDir string) error {
	tf, err := os.Create(tarFile)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %w", err)
	}
	defer tf.Close()

	// Create a tar writer
	tarWriter := tar.NewWriter(tf)
	defer tarWriter.Close()

	if err := addFiletoTar(tarWriter, roverConfigFile, false); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	os.Chdir(schemasDir)
	defer os.Chdir(cwd)
	for _, subgraphSchemaPath := range paths {
		parts := strings.Split(subgraphSchemaPath, "=")
		path := parts[1]
		if err = addFiletoTar(tarWriter, path, true); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	var paths subgraphSchemaPaths
	var roverConfigFile string
	var tarFile string
	var schemasDir string

	flag.Var(&paths, "subgraph", "Specify <subgraph>=<path/to/schema> for each subgraph")
	flag.StringVar(&roverConfigFile, "config", "", "The rover supergraph config file which has pointers to the subgraph schema files")
	flag.StringVar(&tarFile, "output", "", "The tar file which will contain the supergraph files and the individual subgraph schemas")
	flag.StringVar(&schemasDir, "schemas-dir", "", "The base directory inside the sandbox where the schema files will be available.")
	flag.Parse()

	if err := createSubgraphsTarFile(tarFile, paths, roverConfigFile, schemasDir); err != nil {
		log.Fatalf("Error creating tar file: %v", err)
	}
}
