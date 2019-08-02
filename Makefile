.PHONY: all build test pprof

all: test build

build:
	build/env.sh go get -d
	build/env.sh go build -o jibi-run

test:
	build/env.sh go test jibi/*

pprof:
	build/env.sh go test -cpuprofile cpu.prof -memprofile mem.prof -bench . jibi/*
