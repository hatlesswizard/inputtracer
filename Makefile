.PHONY: build test generate clean

build:
	go build ./cmd/... ./pkg/...

test:
	go test ./pkg/...

generate:
	go run ./cmd/genpatterns -o pkg/sources/php/

clean:
	rm -f genpatterns

all: generate build test
