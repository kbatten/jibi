package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/docopt/docopt-go"

	"jibi/pkg/jibi"
)

func main() {
	var config struct {
		DevStatus     bool   `docopt:"--dev-status"`
		DevMaxTicks   int    `docopt:"--dev-maxticks"`
		DevLogInst    bool   `docopt:"--dev-loginstructions"`
		DevCpuProfile bool   `docopt:"--dev-cpuprofile"`
		Rom           string `docopt:"<rom>"`
	}

	usage := `usage: jibi [options] <rom>
dev options:
  --dev-status           show 1 second status
  --dev-maxticks=TICKS   stop after a number of cpu ticks
  --dev-loginstructions  write jibi.log for every instruction
  --dev-cpuprofile       write cpu.prof for use with pprof`

	opts, err := docopt.ParseDoc(usage)
	if err != nil {
		fmt.Println(err)
		return
	}
	opts.Bind(&config)

	// start pprof if required
	if config.DevCpuProfile == true {
		f, err := os.Create("cpu.prof")
		if err != nil {
			fmt.Println(err)
			return
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// load Rom
	rom, err := jibi.ReadRomFile(config.Rom)
	if err != nil {
		fmt.Println(err)
		return
	}

	// create jibi Options
	options := jibi.Options{
		Status:   config.DevStatus,
		MaxTicks: config.DevMaxTicks,
		LogInst:  config.DevLogInst,
	}

	// create jibi and run
	gb := jibi.New(rom, options)
	gb.Run()
}
