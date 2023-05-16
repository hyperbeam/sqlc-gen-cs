package core

import (
	plugin "github.com/tabbed/sqlc-go/codegen"
	sdk "github.com/tabbed/sqlc-go/sdk"
)

func PostgresType(req *plugin.CodeGenRequest, col *plugin.Column, conf *Config) string {
	var csType string
	columnType := sdk.DataType(col.Type)

	switch columnType {
	case "serial", "serial4", "pg_catalog.serial4",
		"integer", "int", "int4", "pg_catalog.int4":
		csType = "int"

	case "bigserial", "serial8", "pg_catalog.serial8",
		"bigint", "int8", "pg_catalog.int8":
		csType = "long"

	case "smallserial", "serial2", "pg_catalog.serial2",
		"smallint", "int2", "pg_catalog.int2":
		csType = "short"

	case "real", "float4", "pg_catalog.float4":
		csType = "float"

	case "float", "double precision", "float8", "pg_catalog.float8":
		csType = "double"

	case "numeric", "pg_catalog.numeric", "money":
		csType = "decimal"

	case "json", "jsonb", "xml",
		"text", "pg_catalog.varchar",
		"pg_catalog.bpchar", "string", "citext",
		"character varying", "character":
		csType = "string"

	case "boolean", "bool", "pg_catalog.bool":
		csType = "bool"

	case "bytea", "blob", "pg_catalog.bytea":
		csType = "byte[]"

	case "date", "pg_catalog.timestamp", "pg_catalog.timestamptz", "timestamptz":
		csType = "DateTime"

	case "pg_catalog.time":
		csType = "TimeSpan"

	case "pg_catalog.timetz":
		csType = "DateTimeOffset"

	case "interval", "pg_catalog.interval":
		csType = "NpgsqlTypes.NpgsqlInterval"

	case "uuid":
		csType = "Guid"

	case "inet":
		csType = "System.Net.IPAddress"

	case "cidr":
		csType = "(System.Net.IPAddress, int)"

	case "macaddr", "macaddr8":
		csType = "System.Net.NetworkInformation.PhysicalAddress"

	case "tsquery":
		csType = "NpgsqlTypes.NpgsqlTsQuery"

	case "tsvector":
		csType = "NpgsqlTypes.NpgsqlTsVector"

	case "ltree", "lquery", "ltxtquery":
		// This module implements a data type ltree for representing labels
		// of data stored in a hierarchical tree-like structure. Extensive
		// facilities for searching through label trees are provided.
		//
		// https://www.postgresql.org/docs/current/ltree.html
		csType = "string"

	case "daterange", "tstzrange", "tsrange":
		csType = "NpgsqlTypes.NpgsqlRange<DateTime>"

	case "tsmultirange", "tstzmultirange", "datemultirange":
		csType = "NpgsqlTypes.NpgsqlRange<DateTime>[]"

	case "numrange":
		csType = "NpgsqlTypes.NpgsqlRange<Decimal>"

	case "nummultirange":
		csType = "NpgsqlTypes.NpgsqlRange<Decimal>[]"

	case "int4range":
		csType = "NpgsqlTypes.NpgsqlRange<int>"

	case "int4multirange":
		csType = "NpgsqlTypes.NpgsqlRange<int>"

	case "int8range":
		csType = "NpgsqlTypes.NpgsqlRange<long>"

	case "int8multirange":
		csType = "NpgsqlTypes.NpgsqlRange<long>[]"

	case "hstore":
		csType = "System.Collections.Generic.Dictionary<string, string>"

	case "bit", "varbit", "pg_catalog.bit", "pg_catalog.varbit":
		csType = "System.Collection.BitArray"

	case "box":
		csType = "NpgsqlTypes.NpgsqlBox"

	case "cid", "oid", "xid":
		csType = "uint"

	case "circle":
		csType = "NpgsqlTypes.NpgsqlCircle"
	case "line":
		csType = "NpgsqlTypes.NpgsqlLine"
	case "lseg":
		csType = "NpgsqlTypes.NpgsqlLSeg"
	case "path":
		csType = "NpgsqlTypes.NpgsqlPath"
	case "point":
		csType = "NpgsqlTypes.NpgsqlPoint"
	case "polygon":
		csType = "NpgsqlTypes.NpgsqlPolygon"
	case "void":
		csType = "object"
	case "any":
		csType = "object"
	}

	if col.IsArray {
		return csType + "[]"
	} else if !col.NotNull && conf.EmitNullOperators {
		return csType + "?"
	} else {
		return csType
	}
}
