package csharp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"text/template"

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

func Generate(ctx context.Context, req *plugin.Request) (*plugin.Response, error) {
	var conf core.Config
	if len(req.PluginOptions) > 0 {
		if err := json.Unmarshal(req.PluginOptions, &conf); err != nil {
			return nil, err
		}
	}

	//enums := core.BuildEnums(req)
	classes := core.BuildClasses(req)
	queries, err := core.BuildQueries(req, conf, classes)
	if err != nil {
		return nil, err
	}

	tctx := TemplateCtx{
		EmitAsync:    conf.EmitAsync,
		SqlcVersion:  req.SqlcVersion,
		CsGenVersion: version,
		Namespace:    conf.Namespace,
	}

	tmpl := template.Must(template.New("table").
		ParseFS(
			templates,
			"templates/*.tmpl",
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
	if err := execute(modelName, "modelsFile"); err != nil {
		return nil, err
	}

	files := map[string]struct{}{}
	for _, gq := range queries {
		files[gq.SourceName] = struct{}{}
	}

	for source := range files {
		if err := execute(source, "queriesFile"); err != nil {
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
