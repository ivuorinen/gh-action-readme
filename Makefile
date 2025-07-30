.PHONY: test lint run example clean readme config-verify

all: test lint

test:
	go test ./...

lint:
	golangci-lint run || true

config-verify:
	golangci-lint config verify --verbose

run:
	go run .

example:
	go run . gen --config config.yaml --output-format=md

readme:
	go run . gen --config config.yaml --output-format=md

clean:
	rm -rf dist/

