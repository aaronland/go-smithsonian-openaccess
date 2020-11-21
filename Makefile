cli:
	go build -mod vendor -o bin/clone cmd/clone/main.go
	go build -mod vendor -o bin/emit cmd/emit/main.go
	go build -mod vendor -o bin/findingaid cmd/findingaid/main.go
	go build -mod vendor -o bin/location cmd/location/main.go
	go build -mod vendor -o bin/placename cmd/placename/main.go
