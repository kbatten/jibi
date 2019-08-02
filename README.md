jibi
====

A golang gameboy emulator.

## About

Currently [boots up the built in bios](http://youtu.be/hfgAkOZB4jU).


## Building

# Makefile

make all

# Manual

build/env.sh go get
build/env.sh go build -o jibi-run


## Profile

./jibi-run --dev-cpuprofile --dev-quick <rom>
go tool pprof -http 0.0.0.0:9999 cpu.prof
