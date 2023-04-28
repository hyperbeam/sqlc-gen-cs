package core

import (
	"fmt"
	"sort"
	"strings"

	plugin "github.com/tabbed/sqlc-go/codegen"
	"github.com/tabbed/sqlc-go/metadata"
	"github.com/tabbed/sqlc-go/sdk"
)

// Column tagged with an ID for matching parameters used multiple times in queries
type codeColumn struct {
	id int
	*plugin.Column
}

func BuildEnums(req *plugin.CodeGenRequest) []Enum {
	var enums []Enum

	for _, schema := range req.Catalog.Schemas {
		if schema.Name == "pg_catalog" || schema.Name == "information_schema" {
			continue
		}

		for _, enum := range schema.Enums {
			var enumName string
			if schema.Name == req.Catalog.DefaultSchema {
				enumName = enum.Name
			} else {
				enumName = schema.Name + "_" + enum.Name
			}

			e := Enum{
				Name:    ClassName(enumName, req.Settings),
				Comment: enum.Comment,
			}

			seen := make(map[string]struct{}, len(enum.Vals))
			for i, v := range enum.Vals {
				value := EnumReplace(v)
				if _, found := seen[value]; found || value == "" {
					value = fmt.Sprintf("%s_%d", value, i)
				}

				e.Members = append(e.Members, EnumMember{
					Name:        ClassName(value, req.Settings),
					MappedValue: v,
				})

				seen[value] = struct{}{}
			}

			enums = append(enums, e)
		}
	}

	if len(enums) > 0 {
		sort.Slice(enums, func(i, j int) bool { return enums[i].Name < enums[j].Name })
	}

	return enums
}

func BuildClasses(req *plugin.CodeGenRequest) []Class {
	var classes []Class
	for _, schema := range req.Catalog.Schemas {
		if schema.Name == "pg_catalog" || schema.Name == "information_schema" {
			continue
		}
		for _, table := range schema.Tables {
			var tableName string
			if schema.Name == req.Catalog.DefaultSchema {
				tableName = table.Rel.Name
			} else {
				tableName = schema.Name + "_" + table.Rel.Name
			}
			className := tableName

			c := Class{
				Table:   &plugin.Identifier{Schema: schema.Name, Name: table.Rel.Name},
				Name:    ClassName(className, req.Settings),
				Comment: table.Comment,
			}

			for _, column := range table.Columns {
				c.Members = append(c.Members, ClassMember{
					Name:    ClassName(column.Name, req.Settings),
					Type:    CsType(req, column),
					Comment: column.Comment,
				})
			}

			classes = append(classes, c)
		}
	}

	if len(classes) > 0 {
		sort.Slice(classes, func(i, j int) bool { return classes[i].Name < classes[j].Name })
	}

	return classes
}

func BuildQueries(req *plugin.CodeGenRequest, conf Config, classes []Class) ([]Query, error) {
	qs := make([]Query, 0, len(req.Queries))
	for _, query := range req.Queries {
		if query.Name == "" {
			continue
		}
		if query.Cmd == "" {
			continue
		}

		constantName := sdk.Title(query.Name)

		gq := Query{
			Cmd:          query.Cmd,
			constantName: constantName,
			MethodName:   query.Name,
			SourceName:   query.Filename,
			SQL:          query.Text,
			Comments:     query.Comments,
			Table:        query.InsertIntoTable,
		}
		if len(query.Params) == 1 && conf.QueryParamLimit != 0 {
			var cols []codeColumn
			for _, p := range query.Params {
				cols = append(cols, codeColumn{
					id:     int(p.Number),
					Column: p.Column,
				})
			}
			c, err := columnsToClass(req, gq.MethodName+"Params", cols, false)
			if err != nil {
				return nil, err
			}
			gq.Args = QueryValue{
				Emit:  true,
				Name:  "arg",
				Class: c,
			}

			if len(query.Params) <= conf.QueryParamLimit {
				gq.Args.Emit = false
			}
		}

		if len(query.Columns) == 1 {
			c := query.Columns[0]
			name := columnName(c, 0)
			if c.IsFuncCall {
				name = strings.Replace(name, "$", "_", -1)
			}
			gq.Ret = QueryValue{
				Name:   name,
				DBName: name,
				Typ:    CsType(req, c),
			}
		} else if putOutColumns(query) {
			var gs *Class
			var emit bool

			for _, class := range classes {
				if len(class.Members) != len(query.Columns) {
					continue
				}
				same := true
				for i, f := range class.Members {
					c := query.Columns[i]
					sameName := f.Name == ClassName(columnName(c, i), req.Settings)
					sameType := f.Type == CsType(req, c)
					sameTable := sdk.SameTableName(c.Table, class.Table, req.Catalog.DefaultSchema)
					if !sameName || !sameType || !sameTable {
						same = false
					}
				}
				if same {
					gs = &class
					break
				}
			}

			if gs == nil {
				var columns []codeColumn
				for i, c := range query.Columns {
					columns = append(columns, codeColumn{
						id:     i,
						Column: c,
					})
				}
				var err error
				gs, err = columnsToClass(req, gq.MethodName+"Row", columns, true)
				if err != nil {
					return nil, err
				}
				emit = true
			}
			gq.Ret = QueryValue{
				Emit:  emit,
				Name:  "i",
				Class: gs,
			}
		}

		qs = append(qs, gq)
	}

	sort.Slice(qs, func(i, j int) bool { return qs[i].MethodName < qs[j].MethodName })
	return qs, nil
}

func columnName(c *plugin.Column, pos int) string {
	if c.Name != "" {
		return c.Name
	}
	return fmt.Sprintf("column_%d", pos+1)
}

func paramName(p *plugin.Parameter) string {
	if p.Column.Name != "" {
		return argName(p.Column.Name)
	}
	return fmt.Sprintf("dollar_%d", p.Number)
}

func argName(name string) string {
	out := ""
	for i, p := range strings.Split(name, "_") {
		if i == 0 {
			out += strings.ToLower(p)
		} else if p == "id" {
			out += "ID"
		} else {
			out += strings.Title(p)
		}
	}
	return out
}

func putOutColumns(query *plugin.Query) bool {
	if len(query.Columns) > 0 {
		return true
	}
	for _, allowed := range []string{metadata.CmdMany, metadata.CmdOne, metadata.CmdBatchMany} {
		if query.Cmd == allowed {
			return true
		}
	}
	return false
}

func columnsToClass(req *plugin.CodeGenRequest, name string, columns []codeColumn, useID bool) (*Class, error) {
	class := Class{
		Name: name,
	}
	seen := map[string][]int{}
	suffixes := map[int]int{}

	for i, c := range columns {
		colName := columnName(c.Column, i)
		memberName := ClassName(colName, req.Settings)
		baseMemberName := memberName

		suffix := 0
		if o, ok := suffixes[c.id]; ok && useID {
			suffix = o
		} else if v := len(seen[memberName]); v > 0 && !c.IsNamedParam {
			suffix = v + 1
		}

		suffixes[c.id] = suffix
		if suffix > 0 {
			memberName = fmt.Sprintf("%s_%d", memberName, suffix)
		}

		member := ClassMember{
			Name:   memberName,
			DBName: colName,
			Column: c.Column,
			Type:   CsType(req, c.Column),
		}

		class.Members = append(class.Members, member)
		if _, found := seen[baseMemberName]; !found {
			seen[baseMemberName] = []int{i}
		} else {
			seen[baseMemberName] = append(seen[baseMemberName], i)
		}
	}

	for i, member := range class.Members {
		if len(seen[member.Name]) > 1 && member.Type == "object" {
			for _, j := range seen[member.Name] {
				if i == j {
					continue
				}
				otherMember := class.Members[j]
				if otherMember.Type != member.Type {
					member.Type = otherMember.Type
				}
				class.Members[i] = member
			}
		}
	}

	err := checkIncompatibleMemberTypes(class.Members)
	if err != nil {
		return nil, err
	}

	return &class, nil
}

func checkIncompatibleMemberTypes(members []ClassMember) error {
	memberTypes := map[string]string{}
	for _, member := range members {
		if memberType, found := memberTypes[member.Name]; !found {
			memberTypes[member.Name] = member.Type
		} else if member.Type != memberType {
			return fmt.Errorf("named param %s has incompatible types: %s, %s", member.Name, member.Type, memberType)
		}
	}

	return nil
}
