build:
	go build -i -v -ldflags="-X csharp.version=$(git describe --always --long --dirty)" github.com/hyperbeam/sqlc-gen-cs