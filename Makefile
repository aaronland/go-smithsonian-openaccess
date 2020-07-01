cli-tools:
	go build -mod vendor -o bin/emit cmd/emit/main.go
	go build -mod vendor -o bin/findingaid cmd/findingaid/main.go
