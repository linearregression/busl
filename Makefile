test: .PHONY
	env $$(cat .env) go test ./...

web: .PHONY
	env $$(cat .env) go run main.go

.PHONY:
