GO_FILES := $(shell find . -type f -name '*.go' -not -path "./Godeps/*" -not -path "./vendor/*")
GO_PACKAGES := $(shell go list ./... | sed "s/github.com\/heroku\/busl/./" | grep -v "^./vendor/")

travis: tidy test

test: .PHONY
	env $$(cat .env) go test ./...

setup: hooks tidy
	cp .env.sample .env

hooks:
	ln -fs ../../bin/git-pre-commit.sh .git/hooks/pre-commit

precommit: tidy test

tidy: goimports
	./bin/go-version-sync-check.sh
	test -z "$$(goimports -l -d $(GO_FILES) | tee /dev/stderr)"
	go vet $(GO_PACKAGES)

web: .PHONY
	env $$(cat .env) go run cmd/busl/main.go

goimports:
	go get golang.org/x/tools/cmd/goimports

busltee: .PHONY bin/busltee

bin/busltee: .PHONY
	docker build -t heroku/busl:latest .
	docker run --rm -i heroku/busl:latest tar cz bin/busltee | tar x

.PHONY:
