.PHONY: all build test pprof vet

all: test build

build:
	build/env.sh go get -d
	build/env.sh go build -o jibi-run

test:
	build/env.sh go test jibi/*

vet:
	build/env.sh go get -d
	build/env.sh go vet
	build/env.sh go vet jibi/*

pprof:
	build/env.sh go test -cpuprofile cpu.prof -memprofile mem.prof -bench . jibi/*
