ApolloSubgraphSchema = provider(
    fields = {
        "service_name": "Name of the service/dgs",
        "schema_file": "The file containing the schema/typedefs for the service",
        "tenant": "The supergraph tenant to which the dgs belongs to",
    },
    doc = "Provider that exposes relevant info about a Apollo Federation subgraph schema.",
)
