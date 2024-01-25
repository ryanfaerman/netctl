# SQL Queries and Migrations

This package provides the queries used by the dao and
the migrations for the database.

## Migration Definition

A model can be generated using the `mage newMigration` command.

When run, a new numerically indexed migration file is added
to the `internal/sql/migrations` directory.

It must define both the up and down migrations.

## Query Definition

The query definitions are grouped, per model or category,
into their own sql file in the `internal/sql/queries` directory.

They are defined according to the following guidelines:

* The file name should be plural of the category or model name
* The methods defined should be descriptive of the query
* Queries returning many should be plural
