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
	v := newVideo()
	m := newMmu(cart, v)
	c := newCpu(m)

	fmt.Println(cart)
	fmt.Println(v)
	fmt.Println(m)
	fmt.Println(c)
	for { //i := 0; i < 5; i++ {
		t := c.step()
		v.step(t)
		//fmt.Println(c)
		if commandTable[c.inst[0]].t == 0 {
			panic("unknown opcode")
		}
	}
}
