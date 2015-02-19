test: .PHONY
	env $$(cat .env) godep go test ./...

.PHONY:
