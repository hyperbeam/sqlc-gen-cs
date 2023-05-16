package core

import (
	plugin "github.com/tabbed/sqlc-go/codegen"
	sdk "github.com/tabbed/sqlc-go/sdk"
)

func CsType(req *plugin.CodeGenRequest, col *plugin.Column, conf *Config) string {
	for _, oride := range req.Settings.Overrides {
		if oride.CodeType == "" {
			continue
		}

		sameTable := sdk.Matches(oride, col.Table, req.Catalog.DefaultSchema)
		if oride.Column != "" && sdk.MatchString(oride.ColumnName, col.Name) && sameTable {
			return oride.CodeType
		}
	}

	typ := csInnerType(req, col, conf)
	if col.IsArray {
		return typ + "[]"
	}

	return typ
}

func csInnerType(req *plugin.CodeGenRequest, col *plugin.Column, conf *Config) string {
	columnType := sdk.DataType(col.Type)
	notNull := col.NotNull || col.IsArray

	// package overrides have a higher precedence
	for _, oride := range req.Settings.Overrides {
		if oride.CodeType == "" {
			continue
		}
		if oride.DbType != "" && oride.DbType == columnType && oride.Nullable != notNull {
			return oride.CodeType
		}
	}

	switch req.Settings.Engine {
	case "postgresql":
		return PostgresType(req, col, conf)
	default:
		return "object"
	}
}
