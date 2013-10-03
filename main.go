package main

import (
	"fmt"
	docopt "github.com/docopt/docopt.go"

	"github.com/kbatten/jibi/jibi"
)

func main() {
	doc := `usage: jibi [options] <rom>
options:
  --dev-norender  disable rendering
  --dev-quick     run a quick test cycle`
	args, _ := docopt.Parse(doc, nil, true, "", false)

	rom, err := jibi.ReadRomFile(args["<rom>"].(string))
	if err != nil {
		fmt.Println(err)
		return
	}

	options := jibi.Options{
		Render: !args["--dev-norender"].(bool),
		Quick:  args["--dev-quick"].(bool),
	}
	gameboy := jibi.New(rom, options)

	gameboy.Run()
}
