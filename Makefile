VERSION=0.0.1
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

all: check-lastlog

check-lastlog: main.go
	go build $(LDFLAGS) -o check-lastlog main.go

linux: main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o check-lastlog main.go

check:
	go test ./...

fmt:
	go fmt ./...

tag:
	git tag v${VERSION}
	git push origin v${VERSION}
	git push origin master
	goreleaser --rm-dist