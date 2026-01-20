.PHONY: build test run clean install

build:
	go build -o bin/task ./cmd/task

install:
	go install ./cmd/task

test:
	go test -v -race -cover ./...

run:
	go run ./cmd/task

clean:
	rm -rf bin/
