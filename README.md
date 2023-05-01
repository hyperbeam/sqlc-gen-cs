# sqlc-gen-cs: a very early beta C# plugin for [SQLC](https://github.com/kyleconroy/sqlc)

sqlc-gen-cs is a beta plugin for adding C# support via ADO .Net to [SQLC](https://github.com/kyleconroy/sqlc)

**this plugin currently only supports Postgresql and cannot handle Enums! We're looking to expand this support very soon!**

## Getting Started

1. Clone the repo
2. In the repo folder run ``make build``
3. In your C# solution create your SQL files according to [SQLC's getting started](https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html)
4. Change you sqlc.yaml to the following, editing the configuration accordingly
```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries.sql"
    schema: "schema.sql"
    codegen:
    - out: models
      plugin: sqlc-gen-cs
      options:
        namespace: hyperbeam.models
        query_param_limit: 1
        emit_async: true
plugins:
- name: sqlc-gen-cs
  process:
    cmd: ./path/to/sqlc-gen-cs
```
5. Run sqlc generate
6. Enjoy your new CS files (Maybe run ``dotnet format`` on them too)

## Configuration

Currently supported plugin configuration options are:
* Most language agnostic config options from Sqlc [as seen here](https://docs.sqlc.dev/en/latest/reference/config.html) barring engine. If you find any unsupported options open up an issue!
* ``namespace`` - The namespace for the generated files
* ``query_param_limit`` - The amount of parameters to inline in the function declaration before creating a new class. -1 is no limit, 0 is invalid.
* ``emit_async`` - whether or not to emit ``await`` compatible functions. Defaults to sync functions.