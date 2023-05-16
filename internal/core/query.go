package core

import (
	"log"
	"strings"

	plugin "github.com/tabbed/sqlc-go/codegen"
)

type Query struct {
	Cmd          string
	Comments     []string
	MethodName   string
	ConstantName string
	SQL          string
	SourceName   string
	Arg          QueryValue
	Ret          QueryValue

	Table *plugin.Identifier
}

func (q Query) HasArgs() bool {
	return !q.Arg.isEmpty()
}

// QueryValue is the holder for our IO part of the query
// It exists to hold a new class, or an existing one.
type QueryValue struct {
	Emit    bool
	Name    string
	DBName  string
	Class   *Class
	Typ     string
	NotNull bool

	Column *plugin.Column
}

func (v QueryValue) EmitClass() bool {
	return v.Emit
}

func (v QueryValue) IsClass() bool {
	return v.Class != nil
}

func (v QueryValue) isEmpty() bool {
	return v.Typ == "" && v.Name == "" && v.Class == nil
}

func (v QueryValue) Type() string {
	if v.Typ != "" {
		return v.Typ
	}
	if v.Class != nil {
		return v.Class.Name
	}
	panic("no type for QueryValue: " + v.Name)
}

func (v QueryValue) EmitReturnType(emitNull bool) string {
	if !emitNull {
		return v.Type()
	}

	// Return types may always be null
	if strings.HasSuffix(v.Type(), "?") {
		return v.Type()
	} else {
		return v.Type() + "?"
	}
}

func (v QueryValue) Pair() string {
	log.Println("Arg value pair: ", v)
	if v.isEmpty() {
		return ""
	}

	var out []string
	if !v.EmitClass() && v.IsClass() {
		for _, f := range v.Class.Members {
			out = append(out, f.Type+" "+strings.ToLower(f.Name))
		}

		return strings.Join(out, ", ")
	}

	return v.Type() + " " + v.Name
}

func (v QueryValue) UniqueMembers() []ClassMember {
	seen := map[string]struct{}{}
	members := make([]ClassMember, 0, len(v.Class.Members))

	for _, member := range v.Class.Members {
		if _, found := seen[member.Name]; found {
			continue
		}
		seen[member.Name] = struct{}{}
		members = append(members, member)
	}

	return members
}
