# rules_apollo_federation
Bazel rules and tools to compose [Apollo Federation](https://www.apollographql.com/docs/graphos/schema-design/federated-schemas/federation) supergraphs and validate subgraphs.

## Features
- Compose federated supergraphs using pure Bazel build actions which are cacheable.
- Validate changes to subgraphs using a subgraph validator to prevent production breakages.
- Use static subgraph schemas or generate dynamic schemas with your own rules and use them as inputs.
