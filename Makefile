.PHONY: all build test

all: test build

build:
	build/env.sh go get -d
	build/env.sh go build -o jibi-run

test:
	build/env.sh go test jibi/*
