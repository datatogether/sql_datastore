# sql_datastore

<!-- Repo Badges for: Github Project, Slack, License-->

[![GitHub](https://img.shields.io/badge/project-Data_Together-487b57.svg?style=flat-square)](http://github.com/datatogether)
[![Slack](https://img.shields.io/badge/slack-Archivers-b44e88.svg?style=flat-square)](https://archivers-slack.herokuapp.com/)
[![License](https://img.shields.io/github/license/datatogether/sql_datastore.svg)](./LICENSE) 

sql_datastore is an experimental Golang implementation of the [ipfs
datastore interface](https://github.com/ipfs/interface-datastore) for
sql databases. Born out of a somewhat special use case of needing to
be able to store data in a number of different places (with the
datastore interface as a lowest-common-denominator), it remains
somewhat experimental.

## License & Copyright

Copyright (C) <year> Data Together

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU Affero General Public License as published by the Free Software
Foundation, version 3.0.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE.

See the [`LICENSE`](./LICENSE) file for details.

## Getting Involved

We would love involvement from more people! If you notice any errors or would like to submit changes, please see our [Contributing Guidelines](./.github/CONTRIBUTING.md). 

We use GitHub issues for [tracking bugs and feature requests](https://github.com/datatogether/sql_datastore/issues) and Pull Requests (PRs) for [submitting changes](https://github.com/datatogether/sql_datastore/pulls)

## Usage

Include in any golang package with:

`import "github.com/datatogether/sql_datastore"`

Implement a new datastore with:
```go
var DefaultStore = NewDatastore(nil)
```
(This is a *package level* Datastore. Be sure to call `SetDB` before using!

Further technical documentation can be built with `godoc .` or, if your `$GOPATH` and repo structure are set up correctly, with something like `godoc -http=:6060 &` followed by browsing to http://localhost:6060/pkg/github.com/datatogether .

## Development

The goal of this package is not a fully-expressive sql database
operated through the datastore interface. This is not possible, or
even desired. Instead, this package focuses on doing the kinds of
things one would want to do with a key-value datastore, requiring
implementers to provide a standard set of queries and parameters to
glue everything together. Whenever the datastore interface is not
expressive enough, one can always fall back to standard SQL work.

`sql_datastore` reconciles the key-value orientation of the datastore interface
with the tables/relational orientation of SQL databases through the concept of a
"Model". Model is a bit of an unfortunate name, as it implies this package is an
ORM, which isn't a design goal.

The important patterns of this approach are:

    1. The Model interface defines how to get stuff into and out of SQL
    2. All Models that will be interacted with must be "Registered" to the store.
       Registered Models map to a datastore.Key Type.
    3. All Get/Put/Delete/Has/Query to sql_datastore must map to a single Model

The current implementation requires substantial boilerplate code to
implement any new interface. In the future this package could be
expanded to become syntax-aware, accepting a table name & schema
definition for registered models. From here the sql_datastore package
could construct default queries that could be overridden using the
current SQLQuery & SQLParams methods. Before that happens, it's worth
noting that the underlying datastore interface may undergo changes in
the near future.


