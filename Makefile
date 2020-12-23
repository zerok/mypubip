all: bin/mypubip

bin:
	mkdir bin

bin/mypubip: $(shell find . -name '*.go') go.mod go.sum bin
	go build -o bin/mypubip

clean:
	rm -rf bin

test:
	go test ./... -v

.PHONY: all clean test
