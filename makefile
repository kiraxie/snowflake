.PHONY: help
help: ## print help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: benchmark
benchmark: ## run go benchmark
	@go test -bench . -benchmem -benchtime=10000000x

.PHONY: test
test: ## run go test
	@go test . -race -failfast -count=1 -v -cover
