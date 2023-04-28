package main

import (
	"github.com/tabbed/sqlc-go/codegen"

	csharp "github.com/hyperbeam/sqlc-gen-cs/internal"
)

func main() {
	codegen.Run(csharp.Generate)
}
