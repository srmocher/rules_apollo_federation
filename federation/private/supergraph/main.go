// This is a simple tool takes a number of subgraph schema names and their paths
// and a output yaml file that can be used to run `rover supergraph compose` to compose
// a Apollo Federation supergraph locally without any external dependencies.

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
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

type subgraphSchemaPaths []string

func (s *subgraphSchemaPaths) String() string {
	return strings.Join(*s, ",")
}

func (s *subgraphSchemaPaths) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func getApolloConfigFromPaths(paths subgraphSchemaPaths) (*apolloSupergraphConfig, error) {
	var apolloConfig apolloSupergraphConfig
	subgraphs := make(map[string]subgraph)
	for _, p := range paths {
		parts := strings.Split(p, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("%s is invalid argument, need arguments in <subgraph-name>=</path/to/service>", p)
		}
		subgraphName, subgraphSchemaFile := parts[0], parts[1]
		subgraphs[subgraphName] = subgraph{
			RoutingUrl: "routing-url-not-required",
			Schema: subgraphSchema{
				File: subgraphSchemaFile,
			},
		}
	}

	apolloConfig.Subgraphs = subgraphs
	return &apolloConfig, nil
}

func main() {
	var paths subgraphSchemaPaths
	var supergraphConfigFile string
	var federationVersion string

	flag.Var(&paths, "subgraph", "Specify <subgraph>=<path/to/schema> for each subgraph")
	flag.StringVar(&supergraphConfigFile, "output", "", "The output supergraph config file to be generated.")
	flag.StringVar(&federationVersion, "federation-version", "=2.9.0", "The version of apollo federation spec to use.")

	flag.Parse()

	apolloConfig, err := getApolloConfigFromPaths(paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	apolloConfig.FederationVersion = federationVersion

	yamlData, err := yaml.Marshal(&apolloConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("Error marshaling to YAML: %v\n", err))
		os.Exit(1)
	}

	if err = os.WriteFile(supergraphConfigFile, yamlData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("Error writing to YAML file: %v\n", err))
		os.Exit(1)
	}

}
