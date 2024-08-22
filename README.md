# dev

```
go build -race -o supervisord .
```

# release

```
go build -ldflags="-s -w" -o supervisord .
```

# lint

```
golangci-lint run -c .golangci.yml ./...
betteralign -apply ./...
nilaway ./...
deadcode ./...

gofumpt -l -w .
```
