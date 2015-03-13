test: .PHONY
	env $$(cat .env) godep go test ./...

web: .PHONY
	env $$(cat .env) godep go run main.go

.PHONY:
