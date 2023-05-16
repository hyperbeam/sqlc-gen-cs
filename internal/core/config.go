package core

type Config struct {
	Namespace                   string   `json:"namespace"`
	QueryParamLimit             int      `json:"query_param_limit"`
	EmitAsync                   bool     `json:"emit_async"`
	EmitNullOperators           bool     `json:"emit_null_ops"`
	LogFile                     string   `json:"log_file"`
	EmitExactTableNames         bool     `json:"emit_exact_table_names"`
	InflectionExcludeTableNames []string `json:"inflection_exclude_table_names"`
}
