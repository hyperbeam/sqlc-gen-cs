package csharp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	plugin "github.com/tabbed/sqlc-go/codegen"

	"github.com/hyperbeam/sqlc-gen-cs/internal/core"
)

var version string

type TemplateCtx struct {
	EmitAsync     bool
	SqlcVersion   string
	CsGenVersion  string
	Namespace     string
	QueryFileName string
	CodeQueries   []core.Query
	Enums         []core.Enum
	Classes       []core.Class
}

func (t *TemplateCtx) OutputQuery(sourceName string) bool {
	return t.QueryFileName == StripExtension(sourceName)
}

func (t *TemplateCtx) ClassName() {

}

func Generate(ctx context.Context, req *plugin.Request) (*plugin.Response, error) {
	var conf core.Config
	if len(req.PluginOptions) > 0 {
		if err := json.Unmarshal(req.PluginOptions, &conf); err != nil {
			return nil, err
		}
	}

	if conf.LogFile != "" {
		f, err := os.OpenFile(conf.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error opening log file: %v", err)
		}

		defer f.Close()
		log.SetOutput(f)
	}

	log.Println("Beginning generation with config: ", conf)
	//enums := core.BuildEnums(req)
	classes := core.BuildClasses(req)
	queries, err := core.BuildQueries(req, conf, classes)
	log.Println("queries built: ", queries)
	if err != nil {
		return nil, err
	}

	tctx := TemplateCtx{
		EmitAsync:    conf.EmitAsync,
		SqlcVersion:  req.SqlcVersion,
		CsGenVersion: version,
		Namespace:    conf.Namespace,
		Classes:      classes,
	}

	funcMap := template.FuncMap{
		"comment":   DoubleSlashComment,
		"classname": RawClassName,
	}

	tmpl := template.Must(template.New("table").
		Funcs(funcMap).
		ParseFS(
			templates,
			"templates/*/*.tmpl",
		),
	)

	output := map[string]string{}

	execute := func(name, templateName string) error {
		var b bytes.Buffer
		w := bufio.NewWriter(&b)

		tctx.QueryFileName = name
		tctx.CodeQueries = queries
		err := tmpl.ExecuteTemplate(w, templateName, &tctx)
		w.Flush()
		if err != nil {
			return err
		}

		code := b.Bytes()
		// TODO: implement auto formatting using dotnet tools

		output[name] = string(code)
		return nil
	}

	modelName := "Models"
	log.Println("Creating models for: ", tctx.Classes)
	if err := execute(modelName, "modelsFile"); err != nil {
		return nil, err
	}

	files := map[string]struct{}{}
	for _, gq := range queries {
		files[gq.SourceName] = struct{}{}
	}

	for source := range files {
		name := StripExtension(source)
		if err := execute(name, "queriesFile"); err != nil {
			return nil, err
		}
	}

	resp := plugin.CodeGenResponse{}
	for filename, code := range output {
		resp.Files = append(resp.Files, &plugin.File{
			Name:     filename + ".cs",
			Contents: []byte(code),
		})
	}

	return &resp, nil
}

func DoubleSlashComment(s string) string {
	return "// " + strings.ReplaceAll(s, "\n", "\n// ")
}

func RawClassName(name string) string {
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

func StripExtension(val string) string {
	extension := filepath.Ext(val)
	return val[0 : len(val)-len(extension)]
}
