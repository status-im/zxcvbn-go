package data

// bindata.go embeds the adjacency graphs in data/. Regenerate it with go generate.
//go:generate go run github.com/go-bindata/go-bindata/go-bindata@v3.1.2+incompatible -pkg data -mode 0644 -modtime 1452717629 -o bindata.go data/
//go:generate gofmt -w bindata.go
