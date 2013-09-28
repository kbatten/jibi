package main

import (
	"fmt"
	docopt "github.com/docopt/docopt.go"
)

func main() {
	doc := `usage: go-gboy <rom>`
	args, _ := docopt.Parse(doc, nil, true, "", false)

	rom := readRomFile(args["<rom>"].(string))

	cart := newCartridge(rom)
	c := newCpu(cart)

	fmt.Println(cart)
	fmt.Println(c)
	for { //i := 0; i < 5; i++ {
		c.loop()
		fmt.Println(c)
		if commandTable[c.inst[0]].t == 0 {
			panic("unknown opcode")
		}
	}
}
