build:
	go build -v -o sqlc-gen-cs -ldflags="-X csharp.version=$(git describe --always --long --dirty)" plugin/main.go