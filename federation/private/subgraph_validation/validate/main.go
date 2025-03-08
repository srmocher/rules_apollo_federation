package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
)

func getSchemaPathForSubgraph(subgraph string, tenant string) string {
	return fmt.Sprintf("%s/schemas/%s/schema.sdl", tenant, subgraph)
}

func validateSubgraph(subgraph string, baseSchemasDir string, currentSchemasDir string, tenant string) error {
	tempSchemasDir, err := os.MkdirTemp("", fmt.Sprintf("%s-validation", subgraph))
	if err != nil {
		return err
	}
	fmt.Printf("--- Validating subgraph %s\n", subgraph)
	defer os.Remove(tempSchemasDir)
	err = copy.Copy(baseSchemasDir, tempSchemasDir)
	if err != nil {
		return err
	}

	baseRoverConfig, err := parseRoverConfigFile(filepath.Join(baseSchemasDir, fmt.Sprintf("%s_rover.yaml", tenant)))
	if err != nil {
		return err
	}
	currentRoverConfig, err := parseRoverConfigFile(filepath.Join(currentSchemasDir, fmt.Sprintf("%s_rover.yaml", tenant)))
	if err != nil {
		return err
	}

	schemaComposeDir := tempSchemasDir
	if _, ok := baseRoverConfig.Subgraphs[subgraph]; ok {
		if _, ok := currentRoverConfig.Subgraphs[subgraph]; ok {
			oldSchemaFile := filepath.Join(tempSchemasDir, getSchemaPathForSubgraph(subgraph, tenant))
			newSchemaFile := filepath.Join(currentSchemasDir, getSchemaPathForSubgraph(subgraph, tenant))
			log.Printf("Copying new schema from %s to %s", newSchemaFile, oldSchemaFile)
			if err = copy.Copy(newSchemaFile, oldSchemaFile); err != nil {
				return err
			}
		}
	} else {
		// if this subgraph doesn't exist on the baseline, then it's probably a new subgraph that got added
		// so we just compose with the current set of schemas and graphs including the new one
		schemaComposeDir = currentSchemasDir
	}

	roverConfigFile := filepath.Join(schemaComposeDir, fmt.Sprintf("%s_rover.yaml", tenant))
	roverPath, err := roverCliPath()
	if err != nil {
		return err
	}
	log.Printf("Running rover composition in directory %s with config %s", schemaComposeDir, roverConfigFile)
	roverCmd := exec.Command(roverPath, "supergraph", "compose", "--config", roverConfigFile)
	roverCmd.Dir = schemaComposeDir
	roverCmd.Stdout = io.Discard
	var stderr bytes.Buffer

	roverCmd.Stderr = &stderr
	env := os.Environ()
	env = append(env, "APOLLO_ELV2_LICENSE=accept")
	roverCmd.Env = env
	if err = roverCmd.Run(); err != nil {
		return fmt.Errorf("subgraph validation failed for %s. The composition error is \n%s", subgraph, stderr.String())
	}

	return nil
}

func validateSubgraphs(sg []string, baseSchemasDir string, currentSchemasDir string, tenant string) map[string]error {
	subgraphValidationErrors := make(map[string]error)
	for _, s := range sg {
		if err := validateSubgraph(s, baseSchemasDir, currentSchemasDir, tenant); err != nil {
			subgraphValidationErrors[s] = err
		}
	}

	return subgraphValidationErrors
}

func main() {
	var baseSubgraphsTarFile string
	var currentSubgraphsTarFile string
	var subgraphs string
	var tenant string
	flag.StringVar(&baseSubgraphsTarFile, "base-subgraphs-tar", "", "Path to tar file containing base subgraphs schemas and rover config")
	flag.StringVar(&currentSubgraphsTarFile, "current-subgraphs-tar", "", "Path to tar file containing subgraph schemas and rover config on current commit")
	flag.StringVar(&subgraphs, "subgraphs", "", "A comma separated list of subgraphs that need to be individually validated.")
	flag.StringVar(&tenant, "tenant", "", "The supergraph tenant")
	flag.Parse()

	if tenant != "FIRST_PARTY" && tenant != "ADMIN" {
		log.Fatalf("Invalid tenant %s", tenant)
	}

	if baseSubgraphsTarFile == "" {
		log.Fatalf("base-subgraphs-tar must be specified!")
	}

	if currentSubgraphsTarFile == "" {
		log.Fatalf("current-subgraphs-tar must be specified!")
	}

	s := strings.Split(subgraphs, ",")

	if len(s) == 0 {
		log.Fatalf("Atleast one subgraph must be passed for validation!")
	}

	baseSchemasDir, err := os.MkdirTemp("", "base-schemas")
	defer os.Remove(baseSchemasDir)
	if err != nil {
		log.Fatalf("Error creating temp dir for schemas: %s", err)
	}
	if err := extractTar(baseSubgraphsTarFile, baseSchemasDir); err != nil {
		log.Fatalf("Couldn't extract base subgraphs tar: %v", err)
	}

	currentSubgraphsDir, err := os.MkdirTemp("", "current-schemas")
	defer os.Remove(currentSubgraphsDir)
	if err != nil {
		log.Fatalf("Error creating temp dir for schemas: %s", err)
	}
	if err := extractTar(currentSubgraphsTarFile, currentSubgraphsDir); err != nil {
		log.Fatalf("Couldn't extract current subgraphs tar: %v", err)
	}

	errors := validateSubgraphs(s, baseSchemasDir, currentSubgraphsDir, tenant)
	if len(errors) == 0 {
		fmt.Println("--- Subgraphs validation result: SUCCESS")
		fmt.Printf("Subgraphs validation passed!\n")
	} else {
		fmt.Println("--- Subgraphs validation result: FAILED")
		fmt.Printf("Subgraph validation failed for the following subgraphs!")
		for subgraph, err := range errors {
			fmt.Printf("%s\n", subgraph)
			fmt.Printf("%s\n", err)
		}
		os.Exit(1)
	}
}
