.DEFAULT_GOAL := help
.PHONY : build

run: ## Run code
	@go run main.go

run-bin: ## Run compiled binary
	@cd bin/ && ./client && cd ../

build: ## Build binary
	@mkdir -p bin
	@go build  -o bin/client main.go

test: ## Run tests
	@go test -v -race ./... -coverprofile=coverage.out

help: ## Show commands availables
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

