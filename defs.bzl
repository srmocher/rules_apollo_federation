load("//federation/private:supergraph.bzl", _apollo_supergraph = "apollo_supergraph")
load("//federation/private:subgraph.bzl", _subgraph_from_schema = "subgraph_from_schema")

apollo_supergraph = _apollo_supergraph
subgraph_from_schema = _subgraph_from_schema