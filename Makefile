.PHONY: all build

all: build

build:
	build/env.sh go get -d
	build/env.sh go build -o jibi-run
