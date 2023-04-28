package core

import (
	"strings"
	"unicode"
	"unicode/utf8"

	plugin "github.com/tabbed/sqlc-go/codegen"
)

type ClassMember struct {
	Name    string
	DBName  string
	Type    string
	Comment string
	Column  *plugin.Column
}

type Class struct {
	Table   *plugin.Identifier
	Name    string
	Members []ClassMember
	Comment string
}

func ClassName(name string, settings *plugin.Settings) string {
	if rename := settings.Rename[name]; rename != "" {
		return rename
	}

	out := ""
	for _, p := range strings.Split(name, "_") {
		if p == "id" {
			out += "ID"
		} else {
			out += strings.Title(p)
		}
	}

	r, _ := utf8.DecodeRuneInString(out)
	if unicode.IsDigit(r) {
		return "_" + out
	} else {
		return out
	}
}
