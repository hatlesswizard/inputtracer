.PHONY: build test coverage fmt-check generate clean

build:
	go build ./cmd/... ./pkg/...

test:
	go test -race ./pkg/...

coverage:
	go test -coverprofile=coverage.out ./pkg/...
	go tool cover -func=coverage.out

fmt-check:
	@gofmt -l $(shell find ./pkg -name '*.go') | grep . && exit 1 || true

generate:
	go run ./cmd/genpatterns -o pkg/sources/php/

clean:
	rm -f genpatterns coverage.out

all: generate build test
