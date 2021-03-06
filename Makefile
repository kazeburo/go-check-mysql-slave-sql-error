VERSION=0.0.8
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"
GO111MODULE=on

all: check-mysql-slave-sql-error

.PHONY: check-mysql-slave-sql-error

check-mysql-slave-sql-error: check-mysql-slave-sql-error.go
	go build $(LDFLAGS) -o check-mysql-slave-sql-error

linux: check-mysql-slave-sql-error.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o check-mysql-slave-sql-error

check:
	go test ./...

fmt:
	go fmt ./...

tag:
	git tag v${VERSION}
	git push origin v${VERSION}
	git push origin master
