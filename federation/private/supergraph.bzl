load(":providers.bzl", "ApolloSubgraphSchema")

# Creates an action that creates the config file required for `rover supergraph compose`
def _rover_config_compose(ctx):
    args = ctx.actions.args()
    rover_config = ctx.actions.declare_file("{}_rover.yaml".format(ctx.attr.tenant))

    subgraph_args, inputs = _subgraph_args_and_inputs(ctx)
    args.add_all(subgraph_args, before_each = "--subgraph")
    args.add("--output", rover_config.path)
    ctx.actions.run(
        executable = ctx.executable._supergraph_setup_tool,
        inputs = depset(inputs),
        outputs = [rover_config],
        arguments = [args],
        progress_message = "Generating Rover Composition config...",
    )

    return rover_config

def _subgraph_args_and_inputs(ctx, use_short_path = True):
    inputs = []
    subgraph_args = []
    for schema_file_target in ctx.attr.subgraph_schemas:
        if ApolloSubgraphSchema not in schema_file_target:
            fail("Not a valid dgs schema target: {}".format(schema_file_target))

        schema_file = schema_file_target[ApolloSubgraphSchema].schema_file
        service_name = schema_file_target[ApolloSubgraphSchema].service_name
        schema_path = schema_file.path
        if use_short_path:
            schema_path = schema_file.short_path
        subgraph_args.append("{}={}".format(service_name, schema_path))
        inputs.append(schema_file)

    return subgraph_args, inputs

def _validate_subgraph_schemas(ctx):
    services = {}
    for schema_target in ctx.attr.subgraph_schemas:
        if schema_target[ApolloSubgraphSchema].service_name in services:
            fail("Multiple subgraph schemas declared for {}: {}, {}. Make sure only one is declared so that composition can happen".format(
                schema_target[ApolloSubgraphSchema].service_name,
                schema_target,
                services[schema_target[ApolloSubgraphSchema].service_name],
            ))
        else:
            services[schema_target[ApolloSubgraphSchema].service_name] = schema_target

def _tar_schema_and_config_files(ctx, rover_config_file):
    tar_file = ctx.actions.declare_file("{}_subgraph_schemas.tar.gz".format(ctx.attr.tenant))
    args = ctx.actions.args()
    subgraph_args, schemas = _subgraph_args_and_inputs(ctx)
    inputs = [rover_config_file] + schemas
    args.add_all(subgraph_args, before_each = "--subgraph")
    args.add("--config", rover_config_file.path)
    args.add("--output", tar_file.path)
    args.add("--schemas-dir", ctx.bin_dir.path)

    ctx.actions.run(
        executable = ctx.executable._tar_subgraphs,
        inputs = inputs,
        outputs = [tar_file],
        arguments = [args],
        mnemonic = "TarSubgraphs",
    )

    return tar_file

def _apollo_supergraph_implementation(ctx):
    supergraph_sdl = ctx.actions.declare_file("{}_supergraph.graphql".format(ctx.attr.tenant))
    _validate_subgraph_schemas(ctx)
    rover_config_file = _rover_config_compose(ctx)
    subgraphs_tar = _tar_schema_and_config_files(ctx, rover_config_file, transformed_schemas)

    inputs = []
    inputs.append(rover_config_file)
    for schema_target in ctx.attr.subgraph_schemas:
        inputs.append(schema_target[ApolloSubgraphSchema].schema_file)

    ctx.actions.run(
        executable = ctx.executable._rover,
        arguments = ["supergraph", "compose", "--config", rover_config_file.path, "--output", supergraph_sdl.path],
        inputs = inputs,
        outputs = [supergraph_sdl],
        progress_message = "Generating supergraph using rover...",
        env = {
            # https://www.apollographql.com/docs/rover/commands/supergraphs#federation-2-elv2-license
            "APOLLO_ELV2_LICENSE": "accept",
        },
        execution_requirements = {
            # rover tries to do a lot of non-hermetic things (access directories outside sandbox)
            # so disable sandboxing for it as it breaks builds on MacOS
            "no-sandbox": "1",
        },
    )

    return [
        DefaultInfo(files = depset([supergraph_sdl])),
        OutputGroupInfo(subgraphs_tar = depset([subgraphs_tar])),
    ]

apollo_supergraph = rule(
    implementation = _apollo_supergraph_implementation,
    attrs = {
        "_rover": attr.label(
            default = Label("//federation/private/supergraph:rover"),
            cfg = "exec",
            executable = True,
        ),
        "tenant": attr.string(
            default = "DEFAULT",
            doc = "The name of the tenant, this is only required if you have multiple supergraphs",
        ),
        "subgraph_schemas": attr.label_list(
            mandatory = True,
            providers = [ApolloSubgraphSchema],
            doc = "A list of subgraph schema targets that are required for composition.",
        ),
        "_supergraph_setup_tool": attr.label(
            default = Label("//federation/private/supergraph_setup"),
            executable = True,
            cfg = "exec",
        ),
        "_tar_subgraphs": attr.label(
            default = Label("//federation/private/subgraph_validation/tar"),
            cfg = "exec",
            executable = True,
        ),
    },
)
