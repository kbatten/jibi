package main

import (
	"fmt"
	docopt "github.com/docopt/docopt.go"

	"github.com/kbatten/jibi/jibi"
)

func main() {
	doc := `usage: jibi [-s | --skipbios] <rom>
options:
  -s --skipbios  start running rom immediately`
	args, _ := docopt.Parse(doc, nil, true, "", false)

	rom, err := jibi.ReadRomFile(args["<rom>"].(string))
	if err != nil {
		fmt.Println(err)
		return
	}

	options := jibi.Options{
		Skipbios: args["--skipbios"].(bool),
	}
	gameboy := jibi.New(rom, options)

	gameboy.Run()
}
