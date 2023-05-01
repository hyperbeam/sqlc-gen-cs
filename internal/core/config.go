package core

type Config struct {
	Namespace       string `json:"namespace"`
	QueryParamLimit int    `json:"query_param_limit"`
	EmitAsync       bool   `json:"emit_async"`
	LogFile         string `json:"log_file"`
}
