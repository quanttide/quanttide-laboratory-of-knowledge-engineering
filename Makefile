.PHONY: build test lint clean

build:
	go build ./...

test:
	go test ./... -v

lint:
	go vet ./...
	@which staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

clean:
	rm -rf parsed/ triples.jsonl errors.jsonl generated/ sample.md
	go clean -cache

extractor:
	go run ./cmd/extractor -input ./docs -output ./parsed
