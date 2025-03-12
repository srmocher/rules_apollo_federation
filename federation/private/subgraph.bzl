load(":providers.bzl", "ApolloSubgraphSchema")
def _subgraph_from_schema_file_impl(ctx):
    return [
        ApolloSubgraphSchema(
            subgraph_name = ctx.attr.subgraph_name,
            schema = ctx.file.schema,
            tenant = ctx.attr.tenant,
        )
    ]


subgraph_from_schema = rule(
    implementation = _subgraph_from_schema_file_impl,
    attrs = {
        "schema": attr.label(
            doc = "The schema file for the subgraph.",
            mandatory = True,
            allow_single_file = True,
        ),
        "subgraph_name": attr.string(
            doc = "The name of the subgraph in the federated supergraph.",
            mandatory = True,
        ),
        "tenant": attr.string(
            doc = "An optional tenant, useful if you have multiple supergraphs",
            default = "DEFAULT",
        )
    }
)