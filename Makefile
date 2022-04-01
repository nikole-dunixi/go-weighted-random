.PHONY: all
all: test benchmark

.PHONY: test
test:
	go test -v --count=1 ./...

.PHONY: benchmark
benchmark:
	go test -bench=./...
