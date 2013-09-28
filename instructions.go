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
	command{"LD BC, nn", 0x01, 2, 12, func(c *cpu) {
		c.c.set(c.inst[1])
		c.b.set(c.inst[2])
	}},
	command{"", 0x02, 0, 0, func(*cpu) {}},
	command{"", 0x03, 0, 0, func(*cpu) {}},
	command{"", 0x04, 0, 0, func(*cpu) {}},
	command{"", 0x05, 0, 0, func(*cpu) {}},
	command{"", 0x06, 0, 8, func(*cpu) {}},
	command{"", 0x07, 0, 0, func(*cpu) {}},
	command{"", 0x08, 0, 0, func(*cpu) {}},
	command{"", 0x09, 0, 0, func(*cpu) {}},
	command{"", 0x0A, 0, 0, func(*cpu) {}},
	command{"", 0x0B, 0, 0, func(*cpu) {}},
	command{"", 0x0C, 0, 0, func(*cpu) {}},
	command{"", 0x0D, 0, 0, func(*cpu) {}},
	command{"", 0x0E, 0, 8, func(*cpu) {}},
	command{"", 0x0F, 0, 0, func(*cpu) {}},
	command{"", 0x10, 0, 0, func(*cpu) {}},
	command{"", 0x11, 0, 0, func(*cpu) {}},
	command{"", 0x12, 0, 0, func(*cpu) {}},
	command{"", 0x13, 0, 0, func(*cpu) {}},
	command{"", 0x14, 0, 0, func(*cpu) {}},
	command{"", 0x15, 0, 0, func(*cpu) {}},
	command{"", 0x16, 0, 0, func(*cpu) {}},
	command{"", 0x17, 0, 0, func(*cpu) {}},
	command{"", 0x18, 0, 0, func(*cpu) {}},
	command{"", 0x19, 0, 0, func(*cpu) {}},
	command{"", 0x1A, 0, 0, func(*cpu) {}},
	command{"", 0x1B, 0, 0, func(*cpu) {}},
	command{"", 0x1C, 0, 0, func(*cpu) {}},
	command{"", 0x1D, 0, 0, func(*cpu) {}},
	command{"", 0x1E, 0, 0, func(*cpu) {}},
	command{"", 0x1F, 0, 0, func(*cpu) {}},
	command{"JR NZ, *", 0x20, 1, 8, func(c *cpu) {
		if c.f.getFlag(flagZ) == false {
			v := int(c.inst[1])
			if v < 0 {
				v = -v
				c.pc -= register16(v)
				return
			}
			c.pc += register16(v)
		}
	}},
	command{"", 0x21, 0, 0, func(*cpu) {}},
	command{"", 0x22, 0, 0, func(*cpu) {}},
	command{"", 0x23, 0, 0, func(*cpu) {}},
	command{"", 0x24, 0, 0, func(*cpu) {}},
	command{"", 0x25, 0, 0, func(*cpu) {}},
	command{"", 0x26, 0, 0, func(*cpu) {}},
	command{"", 0x27, 0, 0, func(*cpu) {}},
	command{"", 0x28, 0, 0, func(*cpu) {}},
	command{"", 0x29, 0, 0, func(*cpu) {}},
	command{"", 0x2A, 0, 0, func(*cpu) {}},
	command{"", 0x2B, 0, 0, func(*cpu) {}},
	command{"", 0x2C, 0, 0, func(*cpu) {}},
	command{"", 0x2D, 0, 0, func(*cpu) {}},
	command{"", 0x2E, 0, 0, func(*cpu) {}},
	command{"", 0x2F, 0, 0, func(*cpu) {}},
	command{"", 0x30, 0, 0, func(*cpu) {}},
	command{"", 0x31, 0, 0, func(*cpu) {}},
	command{"", 0x32, 0, 0, func(*cpu) {}},
	command{"", 0x33, 0, 0, func(*cpu) {}},
	command{"", 0x34, 0, 0, func(*cpu) {}},
	command{"", 0x35, 0, 0, func(*cpu) {}},
	command{"", 0x36, 0, 0, func(*cpu) {}},
	command{"", 0x37, 0, 0, func(*cpu) {}},
	command{"", 0x38, 0, 0, func(*cpu) {}},
	command{"", 0x39, 0, 0, func(*cpu) {}},
	command{"", 0x3A, 0, 0, func(*cpu) {}},
	command{"", 0x3B, 0, 0, func(*cpu) {}},
	command{"", 0x3C, 0, 0, func(*cpu) {}},
	command{"", 0x3D, 0, 0, func(*cpu) {}},
	command{"LD A, #", 0x3E, 1, 8, func(c *cpu) {
		c.a.set(c.inst[1])
	}},
	command{"", 0x3F, 0, 0, func(*cpu) {}},
	command{"", 0x40, 0, 0, func(*cpu) {}},
	command{"", 0x41, 0, 0, func(*cpu) {}},
	command{"", 0x42, 0, 0, func(*cpu) {}},
	command{"", 0x43, 0, 0, func(*cpu) {}},
	command{"", 0x44, 0, 0, func(*cpu) {}},
	command{"", 0x45, 0, 0, func(*cpu) {}},
	command{"", 0x46, 0, 0, func(*cpu) {}},
	command{"", 0x47, 0, 0, func(*cpu) {}},
	command{"", 0x48, 0, 0, func(*cpu) {}},
	command{"", 0x49, 0, 0, func(*cpu) {}},
	command{"", 0x4A, 0, 0, func(*cpu) {}},
	command{"", 0x4B, 0, 0, func(*cpu) {}},
	command{"", 0x4C, 0, 0, func(*cpu) {}},
	command{"", 0x4D, 0, 0, func(*cpu) {}},
	command{"", 0x4E, 0, 0, func(*cpu) {}},
	command{"", 0x4F, 0, 0, func(*cpu) {}},
	command{"", 0x50, 0, 0, func(*cpu) {}},
	command{"", 0x51, 0, 0, func(*cpu) {}},
	command{"", 0x52, 0, 0, func(*cpu) {}},
	command{"", 0x53, 0, 0, func(*cpu) {}},
	command{"", 0x54, 0, 0, func(*cpu) {}},
	command{"", 0x55, 0, 0, func(*cpu) {}},
	command{"", 0x56, 0, 0, func(*cpu) {}},
	command{"", 0x57, 0, 0, func(*cpu) {}},
	command{"", 0x58, 0, 0, func(*cpu) {}},
	command{"", 0x59, 0, 0, func(*cpu) {}},
	command{"", 0x5A, 0, 0, func(*cpu) {}},
	command{"", 0x5B, 0, 0, func(*cpu) {}},
	command{"", 0x5C, 0, 0, func(*cpu) {}},
	command{"", 0x5D, 0, 0, func(*cpu) {}},
	command{"", 0x5E, 0, 0, func(*cpu) {}},
	command{"", 0x5F, 0, 0, func(*cpu) {}},
	command{"", 0x60, 0, 0, func(*cpu) {}},
	command{"", 0x61, 0, 0, func(*cpu) {}},
	command{"", 0x62, 0, 0, func(*cpu) {}},
	command{"", 0x63, 0, 0, func(*cpu) {}},
	command{"", 0x64, 0, 0, func(*cpu) {}},
	command{"", 0x65, 0, 0, func(*cpu) {}},
	command{"", 0x66, 0, 0, func(*cpu) {}},
	command{"", 0x67, 0, 0, func(*cpu) {}},
	command{"", 0x68, 0, 0, func(*cpu) {}},
	command{"", 0x69, 0, 0, func(*cpu) {}},
	command{"", 0x6A, 0, 0, func(*cpu) {}},
	command{"", 0x6B, 0, 0, func(*cpu) {}},
	command{"", 0x6C, 0, 0, func(*cpu) {}},
	command{"", 0x6D, 0, 0, func(*cpu) {}},
	command{"", 0x6E, 0, 0, func(*cpu) {}},
	command{"", 0x6F, 0, 0, func(*cpu) {}},
	command{"", 0x70, 0, 0, func(*cpu) {}},
	command{"", 0x71, 0, 0, func(*cpu) {}},
	command{"", 0x72, 0, 0, func(*cpu) {}},
	command{"", 0x73, 0, 0, func(*cpu) {}},
	command{"", 0x74, 0, 0, func(*cpu) {}},
	command{"", 0x75, 0, 0, func(*cpu) {}},
	command{"", 0x76, 0, 0, func(*cpu) {}},
	command{"", 0x77, 0, 0, func(*cpu) {}},
	command{"", 0x78, 0, 0, func(*cpu) {}},
	command{"", 0x79, 0, 0, func(*cpu) {}},
	command{"", 0x7A, 0, 0, func(*cpu) {}},
	command{"", 0x7B, 0, 0, func(*cpu) {}},
	command{"", 0x7C, 0, 0, func(*cpu) {}},
	command{"", 0x7D, 0, 0, func(*cpu) {}},
	command{"", 0x7E, 0, 0, func(*cpu) {}},
	command{"", 0x7F, 0, 0, func(*cpu) {}},
	command{"", 0x80, 0, 0, func(*cpu) {}},
	command{"", 0x81, 0, 0, func(*cpu) {}},
	command{"", 0x82, 0, 0, func(*cpu) {}},
	command{"", 0x83, 0, 0, func(*cpu) {}},
	command{"", 0x84, 0, 0, func(*cpu) {}},
	command{"ADD A, L", 0x85, 0, 4, func(c *cpu) {
		c.a.set(c.add(c.a.get(), c.l.get()))
	}},
	command{"", 0x86, 0, 0, func(*cpu) {}},
	command{"", 0x87, 0, 0, func(*cpu) {}},
	command{"", 0x88, 0, 0, func(*cpu) {}},
	command{"", 0x89, 0, 0, func(*cpu) {}},
	command{"", 0x8A, 0, 0, func(*cpu) {}},
	command{"", 0x8B, 0, 0, func(*cpu) {}},
	command{"", 0x8C, 0, 0, func(*cpu) {}},
	command{"", 0x8D, 0, 0, func(*cpu) {}},
	command{"", 0x8E, 0, 0, func(*cpu) {}},
	command{"", 0x8F, 0, 0, func(*cpu) {}},
	command{"", 0x90, 0, 0, func(*cpu) {}},
	command{"", 0x91, 0, 0, func(*cpu) {}},
	command{"", 0x92, 0, 0, func(*cpu) {}},
	command{"", 0x93, 0, 0, func(*cpu) {}},
	command{"", 0x94, 0, 0, func(*cpu) {}},
	command{"", 0x95, 0, 0, func(*cpu) {}},
	command{"", 0x96, 0, 0, func(*cpu) {}},
	command{"", 0x97, 0, 0, func(*cpu) {}},
	command{"", 0x98, 0, 0, func(*cpu) {}},
	command{"", 0x99, 0, 0, func(*cpu) {}},
	command{"", 0x9A, 0, 0, func(*cpu) {}},
	command{"", 0x9B, 0, 0, func(*cpu) {}},
	command{"", 0x9C, 0, 0, func(*cpu) {}},
	command{"", 0x9D, 0, 0, func(*cpu) {}},
	command{"", 0x9E, 0, 0, func(*cpu) {}},
	command{"", 0x9F, 0, 0, func(*cpu) {}},
	command{"", 0xA0, 0, 0, func(*cpu) {}},
	command{"", 0xA1, 0, 0, func(*cpu) {}},
	command{"", 0xA2, 0, 0, func(*cpu) {}},
	command{"", 0xA3, 0, 0, func(*cpu) {}},
	command{"", 0xA4, 0, 0, func(*cpu) {}},
	command{"", 0xA5, 0, 0, func(*cpu) {}},
	command{"", 0xA6, 0, 0, func(*cpu) {}},
	command{"", 0xA7, 0, 0, func(*cpu) {}},
	command{"", 0xA8, 0, 0, func(*cpu) {}},
	command{"", 0xA9, 0, 0, func(*cpu) {}},
	command{"", 0xAA, 0, 0, func(*cpu) {}},
	command{"", 0xAB, 0, 0, func(*cpu) {}},
	command{"", 0xAC, 0, 0, func(*cpu) {}},
	command{"", 0xAD, 0, 0, func(*cpu) {}},
	command{"", 0xAE, 0, 0, func(*cpu) {}},
	command{"XOR A", 0xAF, 0, 4, func(c *cpu) {
		c.a.set(c.xor(c.a.get(), c.a.get()))
	}},
	command{"", 0xB0, 0, 0, func(*cpu) {}},
	command{"", 0xB1, 0, 0, func(*cpu) {}},
	command{"", 0xB2, 0, 0, func(*cpu) {}},
	command{"", 0xB3, 0, 0, func(*cpu) {}},
	command{"", 0xB4, 0, 0, func(*cpu) {}},
	command{"", 0xB5, 0, 0, func(*cpu) {}},
	command{"", 0xB6, 0, 0, func(*cpu) {}},
	command{"", 0xB7, 0, 0, func(*cpu) {}},
	command{"", 0xB8, 0, 0, func(*cpu) {}},
	command{"", 0xB9, 0, 0, func(*cpu) {}},
	command{"", 0xBA, 0, 0, func(*cpu) {}},
	command{"", 0xBB, 0, 0, func(*cpu) {}},
	command{"", 0xBC, 0, 0, func(*cpu) {}},
	command{"", 0xBD, 0, 0, func(*cpu) {}},
	command{"", 0xBE, 0, 0, func(*cpu) {}},
	command{"", 0xBF, 0, 0, func(*cpu) {}},
	command{"", 0xC0, 0, 0, func(*cpu) {}},
	command{"", 0xC1, 0, 0, func(*cpu) {}},
	command{"", 0xC2, 0, 0, func(*cpu) {}},
	command{"JP nn", 0xC3, 2, 12, func(c *cpu) {
		c.pc = register16(uint16(c.inst[2])<<8 + uint16(c.inst[1]))
	}},
	command{"", 0xC4, 0, 0, func(*cpu) {}},
	command{"", 0xC5, 0, 0, func(*cpu) {}},
	command{"", 0xC6, 0, 0, func(*cpu) {}},
	command{"", 0xC7, 0, 0, func(*cpu) {}},
	command{"", 0xC8, 0, 0, func(*cpu) {}},
	command{"", 0xC9, 0, 0, func(*cpu) {}},
	command{"", 0xCA, 0, 0, func(*cpu) {}},
	command{"", 0xCB, 0, 0, func(*cpu) {}},
	command{"", 0xCC, 0, 0, func(*cpu) {}},
	command{"", 0xCD, 0, 0, func(*cpu) {}},
	command{"", 0xCE, 0, 0, func(*cpu) {}},
	command{"", 0xCF, 0, 0, func(*cpu) {}},
	command{"", 0xD0, 0, 0, func(*cpu) {}},
	command{"", 0xD1, 0, 0, func(*cpu) {}},
	command{"", 0xD2, 0, 0, func(*cpu) {}},
	command{"", 0xD3, 0, 0, func(*cpu) {}},
	command{"", 0xD4, 0, 0, func(*cpu) {}},
	command{"", 0xD5, 0, 0, func(*cpu) {}},
	command{"", 0xD6, 0, 0, func(*cpu) {}},
	command{"", 0xD7, 0, 0, func(*cpu) {}},
	command{"", 0xD8, 0, 0, func(*cpu) {}},
	command{"", 0xD9, 0, 0, func(*cpu) {}},
	command{"", 0xDA, 0, 0, func(*cpu) {}},
	command{"", 0xDB, 0, 0, func(*cpu) {}},
	command{"", 0xDC, 0, 0, func(*cpu) {}},
	command{"", 0xDD, 0, 0, func(*cpu) {}},
	command{"", 0xDE, 0, 0, func(*cpu) {}},
	command{"", 0xDF, 0, 0, func(*cpu) {}},
	command{"LDH (n), A", 0xE0, 1, 12, func(c *cpu) {
		c.mc.writeByte(Uint16(0xFF00+uint16(c.inst[1])), c.a.get())
	}},
	command{"", 0xE1, 0, 0, func(*cpu) {}},
	command{"", 0xE2, 0, 0, func(*cpu) {}},
	command{"", 0xE3, 0, 0, func(*cpu) {}},
	command{"", 0xE4, 0, 0, func(*cpu) {}},
	command{"", 0xE5, 0, 0, func(*cpu) {}},
	command{"", 0xE6, 0, 0, func(*cpu) {}},
	command{"", 0xE7, 0, 0, func(*cpu) {}},
	command{"", 0xE8, 0, 0, func(*cpu) {}},
	command{"", 0xE9, 0, 0, func(*cpu) {}},
	command{"", 0xEA, 0, 0, func(*cpu) {}},
	command{"", 0xEB, 0, 0, func(*cpu) {}},
	command{"", 0xEC, 0, 0, func(*cpu) {}},
	command{"", 0xED, 0, 0, func(*cpu) {}},
	command{"", 0xEE, 0, 0, func(*cpu) {}},
	command{"", 0xEF, 0, 0, func(*cpu) {}},
	command{"LDH A, (n)", 0xF0, 1, 12, func(c *cpu) {
		c.a.set(c.mc.readByte(Uint16(0xFF00 + uint16(c.inst[1]))))
	}},
	command{"", 0xF1, 0, 0, func(*cpu) {}},
	command{"", 0xF2, 0, 0, func(*cpu) {}},
	command{"DI", 0xF3, 0, 0, func(c *cpu) {
		c.di = true
	}},
	command{"", 0xF4, 0, 0, func(*cpu) {}},
	command{"", 0xF5, 0, 0, func(*cpu) {}},
	command{"", 0xF6, 0, 0, func(*cpu) {}},
	command{"", 0xF7, 0, 0, func(*cpu) {}},
	command{"", 0xF8, 0, 0, func(*cpu) {}},
	command{"", 0xF9, 0, 0, func(*cpu) {}},
	command{"LD A, (nn)", 0xFA, 2, 16, func(c *cpu) {
		nn := uint16(c.inst[2])<<8 + uint16(c.inst[1])
		c.a.set(c.mc.readByte(Uint16(nn)))
	}},
	command{"", 0xFB, 0, 0, func(*cpu) {}},
	command{"", 0xFC, 0, 0, func(*cpu) {}},
	command{"", 0xFD, 0, 0, func(*cpu) {}},
	command{"CP #", 0xFE, 1, 8, func(c *cpu) {
		c.sub(c.a.get(), c.inst[1])
	}},
	command{"", 0xFF, 0, 0, func(*cpu) {}},
}

// holds the instruction currently being fetched
type instruction []uint8

func newInstruction(d ...uint8) instruction {
	inst := make([]uint8, len(d))
	copy(inst, d)
	return inst
}

func (i instruction) String() string {
	if len(i) == 1 {
		opcode := i[0]
		return fmt.Sprintf("< %s [ 0x%02X ] >", commandTable[opcode], i[0])
	}
	if len(i) == 2 {
		opcode := i[0]
		return fmt.Sprintf("< %s [ 0x%02X 0x%02X ] >", commandTable[opcode], i[0], i[1])
	}
	if len(i) == 3 {
		opcode := i[0]
		return fmt.Sprintf("< %s [ 0x%02X 0x%02X 0x%02X ] >", commandTable[opcode], i[0], i[1], i[2])
	}
	return "< >"
}
