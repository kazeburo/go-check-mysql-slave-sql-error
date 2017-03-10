VERSION=0.0.1

all: check-mysql-slave-sql-error

.PHONY: check-mysql-slave-sql-error

bundle:
	dep ensure

check-mysql-slave-sql-error: check-mysql-slave-sql-error.go
	go build -o check-mysql-slave-sql-error

linux: check-mysql-slave-sql-error.go
	GOOS=linux GOARCH=amd64 go build -o check-mysql-slave-sql-error

fmt:
	go fmt ./...

dist:
	git archive --format tgz HEAD -o check-mysql-slave-sql-error-$(VERSION).tar.gz --prefix check-mysql-slave-sql-error-$(VERSION)/

clean:
	rm -rf check-mysql-slave-sql-error check-mysql-slave-sql-error-*.tar.gz

tag:
	git tag v${VERSION}
	git push origin v${VERSION}
	git push origin master

