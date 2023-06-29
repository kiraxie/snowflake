
all:
	@echo -e "Usage:\n\tmake test\n\tmake benchmark"

benchmark:
	@go test -bench . -benchmem -benchtime=10000000x

test:
	@go test . -race -failfast -count=1 -v -cover

.PHONY: all
