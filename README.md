jibi
====

A golang gameboy emulator.

## About

Currently [boots up the built in bios](http://youtu.be/hfgAkOZB4jU).


## Building

go mod init github.com/kbatten/jibi

# Makefile

make all

# Manual

go get
go build -o jibi-run


## Profile

./jibi-run --dev-cpuprofile --dev-quick 5000000 <rom>
go tool pprof -http 0.0.0.0:9999 cpu.prof
