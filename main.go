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
	video := newVideo()
	m := newMmu(cart, video)
	c := newCpu(m)

	fmt.Println(cart)
	fmt.Println(video)
	fmt.Println(m)
	fmt.Println(c)
	/*
		for { //i := 0; i < 5; i++ {
			c.loop()
			fmt.Println(c)
			if commandTable[c.inst[0]].t == 0 {
				panic("unknown opcode")
			}
		}
	*/
}
