BINARY := makethisbetter
MODULE := github.com/makethisbetter/cli
VERSION := 1.0.1
LDFLAGS := -ldflags "-X $(MODULE)/cmd.version=$(VERSION)"

.PHONY: build test clean cross-compile lint

build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test ./...

clean:
	rm -f $(BINARY) $(BINARY)-*

cross-compile:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-linux-arm64 .

lint:
	go vet ./...
