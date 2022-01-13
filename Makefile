.PHONY: all build test pprof vet

all: test build

build:
	go get -d
	go build -o jibi-run

test:
	go get -d
	go test jibi/*

vet:
	go get -d
	go vet
	go vet jibi/*

pprof:
	go test -cpuprofile cpu.prof -memprofile mem.prof -bench . jibi/*
