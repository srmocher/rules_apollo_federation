ApolloSubgraphSchema = provider(
    fields = {
        "subgraph_name": "Name of the subgraph in the federated supergraph",
        "schema": "The file containing the schema/typedefs for the service",
        "tenant": "The supergraph tenant to which the dgs belongs to, only useful if there are multiple supergraphs",
    },
    doc = "Provider that exposes relevant info about a Apollo Federation subgraph schema.",
)
