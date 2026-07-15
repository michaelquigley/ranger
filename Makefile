.PHONY: build test clean

build:
	go build -o vane ./cmd/vane

test:
	go test ./...

clean:
	rm -f vane
