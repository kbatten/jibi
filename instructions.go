package main

import (
	"fmt"
)

type command struct {
	s string
	o uint8 // opcode
	b uint8 // number of immediate bytes
	t uint8 // clock cycles
	f func(*cpu)
}

func (c command) String() string {
	return c.s
}

var commandTable = []command{
	command{"NOP", 0x00, 0, 4, func(*cpu) {}},
	command{"LD BC, nn", 0x01, 2, 4, func(c *cpu) {
		c.b.setWord(c.mc.readWord(c.pc))
		c.pc += 2
	}},
}

// holds the instruction currently being fetched
type instruction []uint8

func newInstruction(d ...uint8) instruction {
	inst := make([]uint8, len(d))
	copy(inst, d)
	return inst
}

func (i instruction) String() string {
	if len(i) > 0 {
		opcode := i[0]
		return fmt.Sprintf("< %v %v >", commandTable[opcode], i[1:])
	}
	return "< >"
}
