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
	0x14: command{"INC D", 0, 4, func(c *cpu) {
		c.d.set(c.inc(c.d.get()))
	}},
	0x15: command{"DEC D", 0, 4, func(c *cpu) {
		c.d.set(c.dec(c.d.get()))
	}},
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
	0x1C: command{"INC E", 0, 4, func(c *cpu) {
		c.e.set(c.inc(c.e.get()))
	}},
	0x1D: command{"DEC E", 0, 4, func(c *cpu) {
		c.e.set(c.dec(c.e.get()))
	}},
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
	0x22: command{"LDI (HL), A", 0, 8, func(c *cpu) {
		c.mc.writeByte(address(c.h.getWord()), c.a.get())
		c.h.setWord(c.h.getWord() + 1)
	}},
	0x23: command{"INC HL", 0, 8, func(c *cpu) {
		c.h.setWord(c.h.getWord() + 1)
	}},
	0x24: command{"INC H", 0, 4, func(c *cpu) {
		c.h.set(c.inc(c.h.get()))
	}},
	0x25: command{"DEC H", 0, 4, func(c *cpu) {
		c.h.set(c.dec(c.h.get()))
	}},
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
	0x28: command{"JR Z, *", 1, 8, func(c *cpu) {
		c.jrF(flagZ, int8(c.inst.p[0]))
	}},
	0x29: command{"", 0, 0, func(c *cpu) {}},
	0x2A: command{"LDI A, (HL)", 0, 8, func(c *cpu) {
		c.a.set(c.mc.readByte(address(c.h.getWord())))
		c.h.setWord(c.h.getWord() + 1)
	}},
	0x2B: command{"", 0, 0, func(c *cpu) {}},
	0x2C: command{"INC L", 0, 4, func(c *cpu) {
		c.l.set(c.inc(c.l.get()))
	}},
	0x2D: command{"DEC L", 0, 4, func(c *cpu) {
		c.l.set(c.dec(c.l.get()))
	}},
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
	0x33: command{"", 0, 0, func(c *cpu) {}},
	0x34: command{"INC (HL)", 0, 12, func(c *cpu) {
		v := c.mc.readByte(address(c.h.getWord()))
		v = c.inc(v)
		c.mc.writeByte(address(c.h.getWord()), v)
	}},
	0x35: command{"DEC (HL)", 0, 12, func(c *cpu) {
		v := c.mc.readByte(address(c.h.getWord()))
		v = c.dec(v)
		c.mc.writeByte(address(c.h.getWord()), v)
	}},
	0x36: command{"LD (HL), n", 1, 12, func(c *cpu) {
		c.mc.writeByte(address(c.h.getWord()), c.inst.p[0])
	}},
	0x37: command{"", 0, 0, func(c *cpu) {}},
	0x38: command{"", 0, 0, func(c *cpu) {}},
	0x39: command{"", 0, 0, func(c *cpu) {}},
	0x3A: command{"", 0, 0, func(c *cpu) {}},
	0x3B: command{"", 0, 0, func(c *cpu) {}},
	0x3C: command{"", 0, 0, func(c *cpu) {}},
	0x3D: command{"DEC A", 0, 4, func(c *cpu) {
		c.a.set(c.dec(c.a.get()))
	}},
	0x3E: command{"LD A, #", 1, 8, func(c *cpu) {
		c.a.set(c.inst.p[0])
	}},
	0x3F: command{"", 0, 0, func(c *cpu) {}},
	0x40: command{"LD B, B", 0, 4, func(c *cpu) { panic("") }},
	0x41: command{"LD B, C", 0, 4, func(c *cpu) { panic("") }},
	0x42: command{"LD B, D", 0, 4, func(c *cpu) { panic("") }},
	0x43: command{"LD B, E", 0, 4, func(c *cpu) { panic("") }},
	0x44: command{"LD B, H", 0, 4, func(c *cpu) { panic("") }},
	0x45: command{"LD B, L", 0, 4, func(c *cpu) { panic("") }},
	0x46: command{"LD B, (HL)", 0, 8, func(c *cpu) { panic("") }},
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
	0x57: command{"LD D, A", 0, 4, func(c *cpu) {
		c.d.set(c.a.get())
	}},
	0x58: command{"", 0, 0, func(c *cpu) {}},
	0x59: command{"", 0, 0, func(c *cpu) {}},
	0x5A: command{"", 0, 0, func(c *cpu) {}},
	0x5B: command{"", 0, 0, func(c *cpu) {}},
	0x5C: command{"", 0, 0, func(c *cpu) {}},
	0x5D: command{"", 0, 0, func(c *cpu) {}},
	0x5E: command{"", 0, 0, func(c *cpu) {}},
	0x5F: command{"LD E, A", 0, 4, func(c *cpu) {
		c.e.set(c.a.get())
	}},
	0x60: command{"", 0, 0, func(c *cpu) {}},
	0x61: command{"", 0, 0, func(c *cpu) {}},
	0x62: command{"", 0, 0, func(c *cpu) {}},
	0x63: command{"", 0, 0, func(c *cpu) {}},
	0x64: command{"", 0, 0, func(c *cpu) {}},
	0x65: command{"", 0, 0, func(c *cpu) {}},
	0x66: command{"", 0, 0, func(c *cpu) {}},
	0x67: command{"LD H, A", 0, 4, func(c *cpu) {
		c.h.set(c.a.get())
	}},
	0x68: command{"", 0, 0, func(c *cpu) {}},
	0x69: command{"", 0, 0, func(c *cpu) {}},
	0x6A: command{"", 0, 0, func(c *cpu) {}},
	0x6B: command{"", 0, 0, func(c *cpu) {}},
	0x6C: command{"", 0, 0, func(c *cpu) {}},
	0x6D: command{"", 0, 0, func(c *cpu) {}},
	0x6E: command{"", 0, 0, func(c *cpu) {}},
	0x6F: command{"LD L, A", 0, 4, func(c *cpu) {
		c.l.set(c.a.get())
	}},
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
	0x90: command{"SUB B", 0, 4, func(c *cpu) {
		c.a.set(c.sub(c.a.get(), c.b.get()))
	}},
	0x91: command{"SUB C", 0, 4, func(c *cpu) {
		c.a.set(c.sub(c.a.get(), c.c.get()))
	}},
	0x92: command{"SUB D", 0, 4, func(c *cpu) {
		c.a.set(c.sub(c.a.get(), c.d.get()))
	}},
	0x93: command{"SUB E", 0, 4, func(c *cpu) {
		c.a.set(c.sub(c.a.get(), c.e.get()))
	}},
	0x94: command{"SUB H", 0, 4, func(c *cpu) {
		c.a.set(c.sub(c.a.get(), c.h.get()))
	}},
	0x95: command{"SUB L", 0, 4, func(c *cpu) {
		c.a.set(c.sub(c.a.get(), c.l.get()))
	}},
	0x96: command{"SUB (HL)", 0, 8, func(c *cpu) {
		v := c.mc.readByte(address(c.h.getWord()))
		c.a.set(c.sub(c.a.get(), v))
	}},
	0x97: command{"SUB A", 0, 4, func(c *cpu) {
		c.a.set(c.sub(c.a.get(), c.a.get()))
	}},
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

	0xB8: command{"CP B", 0, 4, func(c *cpu) {
		c.sub(c.a.get(), c.b.get())
	}},
	0xB9: command{"CP C", 0, 4, func(c *cpu) {
		c.sub(c.a.get(), c.c.get())
	}},
	0xBA: command{"CP D", 0, 4, func(c *cpu) {
		c.sub(c.a.get(), c.d.get())
	}},
	0xBB: command{"CP E", 0, 4, func(c *cpu) {
		c.sub(c.a.get(), c.e.get())
	}},
	0xBC: command{"CP H", 0, 4, func(c *cpu) {
		c.sub(c.a.get(), c.h.get())
	}},
	0xBD: command{"CP L", 0, 4, func(c *cpu) {
		c.sub(c.a.get(), c.l.get())
	}},
	0xBE: command{"CP (HL)", 0, 8, func(c *cpu) {
		v := c.mc.readByte(address(c.h.getWord()))
		c.sub(c.a.get(), v)
	}},
	0xBF: command{"CP A", 0, 4, func(c *cpu) {
		c.sub(c.a.get(), c.a.get())
	}},
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
	0xE3: command{"", 0, 0, func(c *cpu) {}},
	0xE4: command{"", 0, 0, func(c *cpu) {}},
	0xE5: command{"", 0, 0, func(c *cpu) {}},
	0xE6: command{"", 0, 0, func(c *cpu) {}},
	0xE7: command{"", 0, 0, func(c *cpu) {}},
	0xE8: command{"", 0, 0, func(c *cpu) {}},
	0xE9: command{"", 0, 0, func(c *cpu) {}},
	0xEA: command{"LD (nn), A", 2, 16, func(c *cpu) {
		c.mc.writeByte(bytesToAddress(c.inst.p[1], c.inst.p[0]), c.a.get())
	}},
	0xEB: command{"", 0, 0, func(c *cpu) {}},
	0xEC: command{"", 0, 0, func(c *cpu) {}},
	0xED: command{"", 0, 0, func(c *cpu) {}},
	0xEE: command{"", 0, 0, func(c *cpu) {}},
	0xEF: command{"", 0, 0, func(c *cpu) {}},
	0xF0: command{"LDH A, (n)", 1, 12, func(c *cpu) {
		c.a.set(c.mc.readByte(address(0xFF00 + uint16(c.inst.p[0]))))
	}},
	0xF1: command{"", 0, 0, func(c *cpu) {}},
	0xF2: command{"LD A, (C)", 0, 8, func(c *cpu) {
		c.a.set(c.mc.readByte(address(0xFF00 + uint16(c.c.get()))))
	}},
	0xF3: command{"DI", 0, 4, func(c *cpu) {
		c.ime = false
	}},
	0xF4: command{"", 0, 0, func(c *cpu) {}},
	0xF5: command{"", 0, 0, func(c *cpu) {}},
	0xF6: command{"", 0, 0, func(c *cpu) {}},
	0xF7: command{"", 0, 0, func(c *cpu) {}},
	0xF8: command{"LDHL SP, n", 1, 12, func(c *cpu) {
		c.h.setWord(c.addWordR(uint16(c.sp), int8(c.inst.p[0])))
		c.f.resetFlag(flagZ)
		c.f.resetFlag(flagN)
	}},
	0xF9: command{"", 0, 0, func(c *cpu) {}},
	0xFA: command{"LD A, (nn)", 2, 16, func(c *cpu) {
		nn := bytesToAddress(c.inst.p[1], c.inst.p[0])
		c.a.set(c.mc.readByte(nn))
	}},
	0xFB: command{"", 0, 0, func(c *cpu) {}},
	0xFC: command{"", 0, 0, func(c *cpu) {}},
	0xFD: command{"", 0, 0, func(c *cpu) {}},
	0xFE: command{"CP #", 1, 8, func(c *cpu) {
		c.sub(c.a.get(), c.inst.p[0])
	}},
	0xFF: command{"", 0, 0, func(*cpu) {}},
}

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
