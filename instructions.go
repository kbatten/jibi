package main

import (
	"fmt"
)

type command struct {
	s string
	b uint8 // number of immediate bytes
	t uint8 // clock cycles
	f func(*cpu)
}

func (c command) String() string {
	return c.s
}

type opcode uint16

func (o opcode) String() string {
	if c, ok := commandTable[o]; ok {
		if len(c.s) > 0 {
			return c.s
		}
	}
	return fmt.Sprintf("0x%02X", uint16(o))
}

var commandTable = map[opcode]command{
	0x00: command{"NOP", 0, 4, func(*cpu) {}},
	0x01: command{"LD BC, nn", 2, 12, func(c *cpu) {
		c.c.set(c.inst.p[0])
		c.b.set(c.inst.p[1])
	}},
	0x02: command{"LD (BC), A", 0, 8, func(c *cpu) {
		c.mc.writeByte(address(c.b.getWord()), c.a.get())
	}},
	0x03: command{"INC BC", 0, 8, func(c *cpu) {
		c.b.setWord(c.b.getWord() + 1)
	}},
	0x04: command{"INC B", 0, 4, func(c *cpu) {
		c.b.set(c.inc(c.b.get()))
	}},
	0x05: command{"DEC B", 0, 4, func(c *cpu) {
		c.b.set(c.dec(c.b.get()))
	}},
	0x06: command{"LD B, #", 1, 8, func(c *cpu) {
		c.b.set(c.inst.p[0])
	}},
	0x07: command{"RLCA", 4, 0, func(c *cpu) {
		panic("")
		//c.a.set(c.rlc(c.a.get()))
	}},
	0x08: command{"LD (nn), SP", 2, 20, func(c *cpu) {
		c.mc.writeWord(bytesToAddress(c.inst.p[1], c.inst.p[0]), uint16(c.sp))
	}},
	0x09: command{"", 0, 0, func(c *cpu) {}},
	0x0A: command{"", 0, 0, func(c *cpu) {}},
	0x0B: command{"DEC BC", 0, 8, func(c *cpu) {
		c.b.setWord(c.b.getWord() - 1)
	}},
	0x0C: command{"INC C", 0, 4, func(c *cpu) {
		c.c.set(c.inc(c.c.get()))
	}},
	0x0D: command{"DEC C", 0, 4, func(c *cpu) {
		c.c.set(c.dec(c.c.get()))
	}},
	0x0E: command{"LD C, #", 1, 8, func(c *cpu) {
		c.c.set(c.inst.p[0])
	}},
	0x0F: command{"", 0, 0, func(c *cpu) {}},
	0x10: command{"", 0, 0, func(c *cpu) {}},
	0x11: command{"LD DE, nn", 2, 12, func(c *cpu) {
		c.d.setWord(bytesToWord(c.inst.p[1], c.inst.p[0]))
	}},
	0x12: command{"LD (DE), A", 0, 8, func(c *cpu) {
		c.mc.writeByte(address(c.d.getWord()), c.a.get())
	}},
	0x13: command{"INC DE", 0, 8, func(c *cpu) {
		c.d.setWord(c.d.getWord() + 1)
	}},
	0x14: command{"", 0, 0, func(c *cpu) {}},
	0x15: command{"", 0, 0, func(c *cpu) {}},
	0x16: command{"LD D, #", 1, 8, func(c *cpu) {
		c.d.set(c.inst.p[0])
	}},
	0x17: command{"RLA", 0, 4, func(c *cpu) {
		c.a.set(c.rl(c.a.get()))
	}},
	0x18: command{"JR n", 1, 8, func(c *cpu) {
		c.jr(int8(c.inst.p[0]))
	}},
	0x19: command{"", 0, 0, func(c *cpu) {}},
	0x1A: command{"LD A, (DE)", 0, 8, func(c *cpu) {
		c.a.set(c.mc.readByte(address(c.d.getWord())))
	}},
	0x1B: command{"", 0, 0, func(c *cpu) {}},
	0x1C: command{"", 0, 0, func(c *cpu) {}},
	0x1D: command{"", 0, 0, func(c *cpu) {}},
	0x1E: command{"LD E, #", 1, 8, func(c *cpu) {
		c.e.set(c.inst.p[0])
	}},
	0x1F: command{"RRA", 0, 4, func(c *cpu) {
		c.a.set(c.rr(c.a.get()))
	}},
	0x20: command{"JR NZ, *", 1, 8, func(c *cpu) {
		c.jrNF(flagZ, int8(c.inst.p[0]))
	}},
	0x21: command{"LD HL, nn", 2, 12, func(c *cpu) {
		c.h.setWord(bytesToWord(c.inst.p[1], c.inst.p[0]))
	}},
	0x22: command{"", 0, 0, func(c *cpu) {}},
	0x23: command{"", 0, 0, func(c *cpu) {}},
	0x24: command{"", 0, 0, func(c *cpu) {}},
	0x25: command{"", 0, 0, func(c *cpu) {}},
	0x26: command{"LD H, #", 1, 8, func(c *cpu) {
		c.h.set(c.inst.p[0])
	}},
	0x27: command{"DAA", 0, 4, func(c *cpu) {
		a := c.a.get()
		if a&0x0F > 9 || c.f.getFlag(flagH) {
			a += 0x06
			c.f.setFlag(flagH)
		}
		if a > 0x9F || c.f.getFlag(flagC) {
			a += 0x60
			c.f.setFlag(flagC)
		}
	}},
	0x28: command{"", 0, 0, func(c *cpu) {}},
	0x29: command{"", 0, 0, func(c *cpu) {}},
	0x2A: command{"LDI A, (HL)", 0, 8, func(c *cpu) {
		c.a.set(c.mc.readByte(address(c.h.getWord())))
		c.h.setWord(c.h.getWord() + 1)
	}},
	0x2B: command{"", 0, 0, func(c *cpu) {}},
	0x2C: command{"", 0, 0, func(c *cpu) {}},
	0x2D: command{"", 0, 0, func(c *cpu) {}},
	0x2E: command{"LD L, #", 1, 8, func(c *cpu) {
		c.l.set(c.inst.p[0])
	}},
	0x2F: command{"", 0, 0, func(c *cpu) {}},
	0x30: command{"", 0, 0, func(c *cpu) {}},
	0x31: command{"LD SP, nn", 2, 12, func(c *cpu) {
		c.sp = register16(bytesToWord(c.inst.p[1], c.inst.p[0]))
	}},
	0x32: command{"LDD (HL), A", 0, 8, func(c *cpu) {
		c.mc.writeByte(address(c.h.getWord()), c.a.get())
		c.h.setWord(c.h.getWord() - 1)
	}},
	0x33: command{"", 0, 0, func(*cpu) {}},
	0x34: command{"", 0, 0, func(*cpu) {}},
	0x35: command{"", 0, 0, func(*cpu) {}},
	0x36: command{"LD (HL), n", 1, 12, func(c *cpu) {
		c.mc.writeByte(address(c.h.getWord()), c.inst.p[0])
	}},
	0x37: command{"", 0, 0, func(*cpu) {}},
	0x38: command{"", 0, 0, func(*cpu) {}},
	0x39: command{"", 0, 0, func(*cpu) {}},
	0x3A: command{"", 0, 0, func(*cpu) {}},
	0x3B: command{"", 0, 0, func(*cpu) {}},
	0x3C: command{"", 0, 0, func(*cpu) {}},
	0x3D: command{"", 0, 0, func(*cpu) {}},
	0x3E: command{"LD A, #", 1, 8, func(c *cpu) {
		c.a.set(c.inst.p[0])
	}},
	0x3F: command{"", 0, 0, func(*cpu) {}},
	0x40: command{"LD B, B", 0, 4, func(*cpu) { panic("") }},
	0x41: command{"LD B, C", 0, 4, func(*cpu) { panic("") }},
	0x42: command{"LD B, D", 0, 4, func(*cpu) { panic("") }},
	0x43: command{"LD B, E", 0, 4, func(*cpu) { panic("") }},
	0x44: command{"LD B, H", 0, 4, func(*cpu) { panic("") }},
	0x45: command{"LD B, L", 0, 4, func(*cpu) { panic("") }},
	0x46: command{"LD B, (HL)", 0, 8, func(*cpu) { panic("") }},
	0x47: command{"LD B, A", 0, 4, func(c *cpu) {
		c.b.set(c.a.get())
	}},
	0x48: command{"", 0, 0, func(c *cpu) {}},
	0x49: command{"", 0, 0, func(c *cpu) {}},
	0x4A: command{"", 0, 0, func(c *cpu) {}},
	0x4B: command{"", 0, 0, func(c *cpu) {}},
	0x4C: command{"", 0, 0, func(c *cpu) {}},
	0x4D: command{"", 0, 0, func(c *cpu) {}},
	0x4E: command{"", 0, 0, func(c *cpu) {}},
	0x4F: command{"LD C, A", 0, 4, func(c *cpu) {
		c.c.set(c.a.get())
	}},
	0x50: command{"", 0, 0, func(c *cpu) {}},
	0x51: command{"", 0, 0, func(c *cpu) {}},
	0x52: command{"", 0, 0, func(c *cpu) {}},
	0x53: command{"", 0, 0, func(c *cpu) {}},
	0x54: command{"", 0, 0, func(c *cpu) {}},
	0x55: command{"", 0, 0, func(c *cpu) {}},
	0x56: command{"", 0, 0, func(c *cpu) {}},
	0x57: command{"", 0, 0, func(c *cpu) {}},
	0x58: command{"", 0, 0, func(c *cpu) {}},
	0x59: command{"", 0, 0, func(c *cpu) {}},
	0x5A: command{"", 0, 0, func(c *cpu) {}},
	0x5B: command{"", 0, 0, func(c *cpu) {}},
	0x5C: command{"", 0, 0, func(c *cpu) {}},
	0x5D: command{"", 0, 0, func(c *cpu) {}},
	0x5E: command{"", 0, 0, func(c *cpu) {}},
	0x5F: command{"", 0, 0, func(c *cpu) {}},
	0x60: command{"", 0, 0, func(c *cpu) {}},
	0x61: command{"", 0, 0, func(c *cpu) {}},
	0x62: command{"", 0, 0, func(c *cpu) {}},
	0x63: command{"", 0, 0, func(c *cpu) {}},
	0x64: command{"", 0, 0, func(c *cpu) {}},
	0x65: command{"", 0, 0, func(c *cpu) {}},
	0x66: command{"", 0, 0, func(c *cpu) {}},
	0x67: command{"", 0, 0, func(c *cpu) {}},
	0x68: command{"", 0, 0, func(c *cpu) {}},
	0x69: command{"", 0, 0, func(c *cpu) {}},
	0x6A: command{"", 0, 0, func(c *cpu) {}},
	0x6B: command{"", 0, 0, func(c *cpu) {}},
	0x6C: command{"", 0, 0, func(c *cpu) {}},
	0x6D: command{"", 0, 0, func(c *cpu) {}},
	0x6E: command{"", 0, 0, func(c *cpu) {}},
	0x6F: command{"", 0, 0, func(c *cpu) {}},
	0x70: command{"", 0, 0, func(c *cpu) {}},
	0x71: command{"", 0, 0, func(c *cpu) {}},
	0x72: command{"", 0, 0, func(c *cpu) {}},
	0x73: command{"LD (HL), E", 0, 8, func(c *cpu) {
		c.mc.writeByte(address(c.h.getWord()), c.e.get())
	}},
	0x74: command{"", 0, 0, func(c *cpu) {}},
	0x75: command{"", 0, 0, func(c *cpu) {}},
	0x76: command{"", 0, 0, func(c *cpu) {}},
	0x77: command{"LD (HL), A", 0, 8, func(c *cpu) {
		c.mc.writeByte(address(c.h.getWord()), c.a.get())
	}},
	0x78: command{"LD A, B", 0, 4, func(c *cpu) {
		c.a.set(c.b.get())
	}},
	0x79: command{"LD A, C", 0, 4, func(c *cpu) {
		c.a.set(c.c.get())
	}},
	0x7A: command{"LD A, D", 0, 4, func(c *cpu) {
		c.a.set(c.d.get())
	}},
	0x7B: command{"LD A, E", 0, 4, func(c *cpu) {
		c.a.set(c.e.get())
	}},
	0x7C: command{"LD A, H", 0, 4, func(c *cpu) {
		c.a.set(c.h.get())
	}},
	0x7D: command{"LD A, L", 0, 4, func(c *cpu) {
		c.a.set(c.l.get())
	}},
	0x7E: command{"LD A, (HL)", 0, 8, func(c *cpu) {
		c.a.set(c.mc.readByte(address(c.h.getWord())))
	}},
	0x7F: command{"LD A, A", 0, 4, func(c *cpu) {
		c.a.set(c.a.get())
	}},
	0x80: command{"ADD A, B", 0, 4, func(c *cpu) {
		c.a.set(c.add(c.a.get(), c.b.get()))
	}},
	0x81: command{"ADD A, C", 0, 4, func(c *cpu) {
		c.a.set(c.add(c.a.get(), c.c.get()))
	}},
	0x82: command{"ADD A, D", 0, 4, func(c *cpu) {
		c.a.set(c.add(c.a.get(), c.d.get()))
	}},
	0x83: command{"ADD A, E", 0, 4, func(c *cpu) {
		c.a.set(c.add(c.a.get(), c.e.get()))
	}},
	0x84: command{"ADD A, H", 0, 4, func(c *cpu) {
		c.a.set(c.add(c.a.get(), c.h.get()))
	}},
	0x85: command{"ADD A, L", 0, 4, func(c *cpu) {
		c.a.set(c.add(c.a.get(), c.l.get()))
	}},
	0x86: command{"ADD A, (HL)", 0, 8, func(c *cpu) {
		c.a.set(c.add(c.a.get(), c.mc.readByte(address(c.h.getWord()))))
	}},
	0x87: command{"ADD A, A", 0, 4, func(c *cpu) {
		c.a.set(c.add(c.a.get(), c.a.get()))
	}},
	0x88: command{"ADC A, B", 0, 4, func(c *cpu) {
		c.a.set(c.adc(c.a.get(), c.b.get()))
	}},
	0x89: command{"ADC A, C", 0, 4, func(c *cpu) {
		c.a.set(c.adc(c.a.get(), c.c.get()))
	}}, 0x8A: command{"ADC A, D", 0, 4, func(c *cpu) {
		c.a.set(c.adc(c.a.get(), c.d.get()))
	}}, 0x8B: command{"ADC A, E", 0, 4, func(c *cpu) {
		c.a.set(c.adc(c.a.get(), c.e.get()))
	}}, 0x8C: command{"ADC A, H", 0, 4, func(c *cpu) {
		c.a.set(c.adc(c.a.get(), c.h.get()))
	}}, 0x8D: command{"ADC A, L", 0, 4, func(c *cpu) {
		c.a.set(c.adc(c.a.get(), c.l.get()))
	}}, 0x8E: command{"ADC A, (HL)", 0, 8, func(c *cpu) {
		c.a.set(c.adc(c.a.get(), c.mc.readByte(address(c.h.getWord()))))
	}}, 0x8F: command{"ADC A, A", 0, 4, func(c *cpu) {
		c.a.set(c.adc(c.a.get(), c.a.get()))
	}},
	0x90: command{"", 0, 0, func(c *cpu) {}},
	0x91: command{"", 0, 0, func(c *cpu) {}},
	0x92: command{"", 0, 0, func(c *cpu) {}},
	0x93: command{"", 0, 0, func(c *cpu) {}},
	0x94: command{"", 0, 0, func(c *cpu) {}},
	0x95: command{"", 0, 0, func(c *cpu) {}},
	0x96: command{"", 0, 0, func(c *cpu) {}},
	0x97: command{"", 0, 0, func(c *cpu) {}},
	0x98: command{"", 0, 0, func(c *cpu) {}},
	0x99: command{"", 0, 0, func(c *cpu) {}},
	0x9A: command{"", 0, 0, func(c *cpu) {}},
	0x9B: command{"", 0, 0, func(c *cpu) {}},
	0x9C: command{"", 0, 0, func(c *cpu) {}},
	0x9D: command{"", 0, 0, func(c *cpu) {}},
	0x9E: command{"", 0, 0, func(c *cpu) {}},
	0x9F: command{"", 0, 0, func(c *cpu) {}},
	0xA0: command{"", 0, 0, func(c *cpu) {}},
	0xA1: command{"", 0, 0, func(c *cpu) {}},
	0xA2: command{"", 0, 0, func(c *cpu) {}},
	0xA3: command{"", 0, 0, func(c *cpu) {}},
	0xA4: command{"AND H", 0, 4, func(c *cpu) {
		c.a.set(c.and(c.a.get(), c.h.get()))
	}},
	0xA5: command{"", 0, 0, func(c *cpu) {}},
	0xA6: command{"", 0, 0, func(c *cpu) {}},
	0xA7: command{"", 0, 0, func(c *cpu) {}},
	0xA8: command{"", 0, 0, func(c *cpu) {}},
	0xA9: command{"", 0, 0, func(c *cpu) {}},
	0xAA: command{"", 0, 0, func(c *cpu) {}},
	0xAB: command{"", 0, 0, func(c *cpu) {}},
	0xAC: command{"", 0, 0, func(c *cpu) {}},
	0xAD: command{"", 0, 0, func(c *cpu) {}},
	0xAE: command{"", 0, 0, func(c *cpu) {}},
	0xAF: command{"XOR A", 0, 4, func(c *cpu) {
		c.a.set(c.xor(c.a.get(), c.a.get()))
	}},
	0xB0: command{"OR B", 0, 4, func(c *cpu) {
		c.a.set(c.or(c.a.get(), c.b.get()))
	}},
	0xB1: command{"OR C", 0, 4, func(c *cpu) {
		c.a.set(c.or(c.a.get(), c.c.get()))
	}},
	0xB2: command{"OR D", 0, 4, func(c *cpu) {
		c.a.set(c.or(c.a.get(), c.d.get()))
	}},
	0xB3: command{"OR E", 0, 4, func(c *cpu) {
		c.a.set(c.or(c.a.get(), c.e.get()))
	}},
	0xB4: command{"OR H", 0, 4, func(c *cpu) {
		c.a.set(c.or(c.a.get(), c.h.get()))
	}},
	0xB5: command{"OR L", 0, 4, func(c *cpu) {
		c.a.set(c.or(c.a.get(), c.l.get()))
	}},
	0xB6: command{"OR (HL)", 0, 8, func(c *cpu) {
		c.a.set(c.or(c.a.get(), c.mc.readByte(address(c.h.getWord()))))
	}},
	0xB7: command{"", 0, 0, func(c *cpu) {}},
	0xB8: command{"", 0, 0, func(c *cpu) {}},
	0xB9: command{"", 0, 0, func(c *cpu) {}},
	0xBA: command{"", 0, 0, func(c *cpu) {}},
	0xBB: command{"", 0, 0, func(c *cpu) {}},
	0xBC: command{"", 0, 0, func(c *cpu) {}},
	0xBD: command{"", 0, 0, func(c *cpu) {}},
	0xBE: command{"", 0, 0, func(c *cpu) {}},
	0xBF: command{"", 0, 0, func(c *cpu) {}},
	0xC0: command{"", 0, 0, func(c *cpu) {}},
	0xC1: command{"POP BC", 0, 12, func(c *cpu) {
		c.b.setWord(c.popWord())
	}},
	0xC2: command{"", 0, 0, func(c *cpu) {}},
	0xC3: command{"JP nn", 2, 12, func(c *cpu) {
		c.jp(bytesToAddress(c.inst.p[1], c.inst.p[0]))
	}},
	0xC4: command{"", 0, 0, func(c *cpu) {}},
	0xC5: command{"PUSH BC", 0, 16, func(c *cpu) {
		c.pushWord(c.b.getWord())
	}},
	0xC6: command{"", 0, 0, func(c *cpu) {}},
	0xC7: command{"", 0, 0, func(c *cpu) {}},
	0xC8: command{"", 0, 0, func(c *cpu) {}},
	0xC9: command{"RET", 0, 8, func(c *cpu) {
		c.jp(address(c.popWord()))
	}},
	0xCA: command{"", 0, 0, func(c *cpu) {}},
	0xCB11: command{"RL C", 0, 8, func(c *cpu) {
		c.c.set(c.rl(c.c.get()))
	}},
	0xCB7C: command{"BIT 7, H", 0, 8, func(c *cpu) {
		c.bit(7, c.h.get())
	}},
	0xCC: command{"CALL Z, nn", 2, 12, func(c *cpu) {
		c.callF(flagZ, bytesToAddress(c.inst.p[1], c.inst.p[0]))
	}},
	0xCD: command{"CALL nn", 2, 12, func(c *cpu) {
		c.call(bytesToAddress(c.inst.p[1], c.inst.p[0]))
	}},
	0xCE: command{"", 0, 0, func(c *cpu) {}},
	0xCF: command{"", 0, 0, func(c *cpu) {}},
	0xD0: command{"", 0, 0, func(c *cpu) {}},
	0xD1: command{"", 0, 0, func(c *cpu) {}},
	0xD2: command{"", 0, 0, func(c *cpu) {}},
	0xD3: command{"", 0, 0, func(c *cpu) {}},
	0xD4: command{"", 0, 0, func(c *cpu) {}},
	0xD5: command{"", 0, 0, func(c *cpu) {}},
	0xD6: command{"", 0, 0, func(c *cpu) {}},
	0xD7: command{"", 0, 0, func(c *cpu) {}},
	0xD8: command{"", 0, 0, func(c *cpu) {}},
	0xD9: command{"", 0, 0, func(c *cpu) {}},
	0xDA: command{"", 0, 0, func(c *cpu) {}},
	0xDB: command{"", 0, 0, func(c *cpu) {}},
	0xDC: command{"", 0, 0, func(c *cpu) {}},
	0xDE: command{"", 0, 0, func(c *cpu) {}},
	0xDF: command{"", 0, 0, func(c *cpu) {}},
	0xE0: command{"LDH (n), A", 1, 12, func(c *cpu) {
		c.mc.writeByte(address(0xFF00+uint16(c.inst.p[0])), c.a.get())
	}},
	0xE1: command{"", 0, 0, func(*cpu) {}},
	0xE2: command{"LD (C), A", 0, 8, func(c *cpu) {
		c.mc.writeByte(address(0xFF00+uint16(c.c.get())), c.a.get())
	}},
} /*
	command{"", 0xE3, 0, 0, func(c *cpu) {}},
	command{"", 0xE4, 0, 0, func(c *cpu) {}},
	command{"", 0xE5, 0, 0, func(c *cpu) {}},
	command{"", 0xE6, 0, 0, func(c *cpu) {}},
	command{"", 0xE7, 0, 0, func(c *cpu) {}},
	command{"", 0xE8, 0, 0, func(c *cpu) {}},
	command{"", 0xE9, 0, 0, func(c *cpu) {}},
	command{"LD (nn), A", 0xEA, 2, 16, func(c *cpu) {
		c.mc.writeByte(bytesToAddress(c.inst[2], c.inst[1]), c.a.get())
	}},
	command{"", 0xEB, 0, 0, func(c *cpu) {}},
	command{"", 0xEC, 0, 0, func(c *cpu) {}},
	command{"", 0xED, 0, 0, func(c *cpu) {}},
	command{"", 0xEE, 0, 0, func(c *cpu) {}},
	command{"", 0xEF, 0, 0, func(c *cpu) {}},
	command{"LDH A, (n)", 0xF0, 1, 12, func(c *cpu) {
		c.a.set(c.mc.readByte(address(0xFF00 + uint16(c.inst[1]))))
	}},
	command{"", 0xF1, 0, 0, func(c *cpu) {}},
	command{"", 0xF2, 0, 0, func(c *cpu) {}},
	command{"DI", 0xF3, 0, 4, func(c *cpu) {
		// write to Interrupt Enable Register
		// TODO
		c.mc.writeByte(address(0xFFFF), 0x03)
	}},
	command{"", 0xF4, 0, 0, func(c *cpu) {}},
	command{"", 0xF5, 0, 0, func(c *cpu) {}},
	command{"", 0xF6, 0, 0, func(c *cpu) {}},
	command{"", 0xF7, 0, 0, func(c *cpu) {}},
	command{"LDHL SP, n", 0xF8, 1, 12, func(c *cpu) {
		c.h.setWord(c.addWordR(uint16(c.sp), int8(c.inst[1])))
		c.f.resetFlag(flagZ)
		c.f.resetFlag(flagN)
	}},
	command{"", 0xF9, 0, 0, func(c *cpu) {}},
	command{"LD A, (nn)", 0xFA, 2, 16, func(c *cpu) {
		nn := bytesToAddress(c.inst[2], c.inst[1])
		c.a.set(c.mc.readByte(nn))
	}},
	command{"", 0xFB, 0, 0, func(c *cpu) {}},
	command{"", 0xFC, 0, 0, func(c *cpu) {}},
	command{"", 0xFD, 0, 0, func(c *cpu) {}},
	command{"CP #", 0xFE, 1, 8, func(c *cpu) {
		c.sub(c.a.get(), c.inst[1])
	}},
	command{"", 0xFF, 0, 0, func(*cpu) {}},
}
*/
// holds the instruction currently being fetched
type instruction struct {
	o opcode
	p []uint8 // params
}

func newInstruction(o opcode, ps ...uint8) instruction {
	p := make([]uint8, len(ps))
	copy(p, ps)
	return instruction{o, p}
}

func (i instruction) String() string {
	ps := ""
	for _, v := range i.p {
		ps += fmt.Sprintf("0x%02X ", v)
	}
	return fmt.Sprintf("%s [ 0x%02X %s]", i.o, uint16(i.o), ps)
}
