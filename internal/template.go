package csharp

import "embed"

//go:embed templates/*
//go:embed templates/*/*
var templates embed.FS
