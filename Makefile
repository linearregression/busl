test: .PHONY
	env $$(cat .env) go test ./...

web: .PHONY
	env $$(cat .env) go run main.go

busltee: .PHONY bin/busltee

bin/busltee: .PHONY
	docker build -t heroku/busl:latest .
	docker run --rm -i heroku/busl:latest tar cz bin/busltee | tar x

.PHONY:
