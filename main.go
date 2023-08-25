package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"strconv"

	"github.com/docopt/docopt-go"

	"jibi/pkg/jibi"
)

func main() {
	doc := `usage: jibi [options] <rom>
dev options:
  --dev-status         show 1 second status
  --dev-norender       disable rendering
  --dev-nokeypad       disable keypad input
  --dev-quick=<ticks>  run a quick test cycle
  --dev-nosquash       only display upper left
  --dev-every          print every exectuted instruction
  --dev-cpuprofile     write cpu.prof for use with pprof`
	args, _ := docopt.Parse(doc, nil, true, "", false)

	rom, err := jibi.ReadRomFile(args["<rom>"].(string))
	if err != nil {
		fmt.Println(err)
		return
	}

	var quick uint64
	if args["--dev-quick"] != nil {
		quick, err = strconv.ParseUint(args["--dev-quick"].(string), 10, 64)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	options := jibi.Options{
		Status: args["--dev-status"].(bool),
		Render: !args["--dev-norender"].(bool),
		Keypad: !args["--dev-nokeypad"].(bool),
		Quick:  quick,
		Squash: !args["--dev-nosquash"].(bool),
		Every:  args["--dev-every"].(bool),
	}

	if args["--dev-cpuprofile"].(bool) {
		f, err := os.Create("cpu.prof")
		if err != nil {
			fmt.Println(err)
			return
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	gameboy := jibi.New(rom, options)

	gameboy.Run()
}
