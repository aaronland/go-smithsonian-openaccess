GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

cli:
	go build -mod $(GOMOD) -o bin/matches cmd/matches/main.go

vuln:
	govulncheck -show verbose ./...
