BIN := broask
GOPATH_BIN := $(shell go env GOPATH)/bin

.PHONY: build install test clean

build:
	go build -o $(BIN) .

install:
	go install ./...

test:
	go test ./...

clean:
	rm -f $(BIN)
