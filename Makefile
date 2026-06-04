GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

TAGS=

vuln:
	govulncheck -show verbose ./...

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/clone cmd/clone/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/walk cmd/walk/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/emit cmd/emit/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/findingaid cmd/findingaid/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/location cmd/location/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/placename cmd/placename/main.go
