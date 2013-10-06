package main

import (
	"fmt"
	docopt "github.com/docopt/docopt.go"

	"github.com/kbatten/jibi/jibi"
)

func main() {
	doc := `usage: jibi [options] <rom>
dev options:
  --dev-norender  disable rendering
  --dev-nokeypad  disable keypad input
  --dev-quick     run a quick test cycle
  --dev-nosquash  only display upper left`
	args, _ := docopt.Parse(doc, nil, true, "", false)

	rom, err := jibi.ReadRomFile(args["<rom>"].(string))
	if err != nil {
		fmt.Println(err)
		return
	}

	options := jibi.Options{
		Render: !args["--dev-norender"].(bool),
		Keypad: !args["--dev-nokeypad"].(bool),
		Quick:  args["--dev-quick"].(bool),
		Squash: !args["--dev-nosquash"].(bool),
	}
	gameboy := jibi.New(rom, options)

	gameboy.Run()
}
