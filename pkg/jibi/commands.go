package jibi

import (
	"fmt"
)

type command struct {
	s string
	b uint8 // number of immediate bytes
	t uint8 // clock cycles
	f func(*Cpu)
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
	0x00: command{"NOP", 0, 4, func(*Cpu) {}},
	0x01: command{"LD BC, nn", 2, 12, func(c *Cpu) {
		c.c.set(c.inst.p[0])
		c.b.set(c.inst.p[1])
	}},
	0x02: command{"LD (BC), A", 0, 8, func(c *Cpu) {
		c.writeByte(c.b.Word(), c.a.Byte())
	}},
	0x03: command{"INC BC", 0, 8, func(c *Cpu) {
		c.b.setWord(c.b.Word() + 1)
	}},
	0x04: command{"INC B", 0, 4, func(c *Cpu) {
		c.b.set(c.inc(c.b.Byte()))
	}},
	0x05: command{"DEC B", 0, 4, func(c *Cpu) {
		c.b.set(c.dec(c.b.Byte()))
	}},
	0x06: command{"LD B, #", 1, 8, func(c *Cpu) {
		c.b.set(c.inst.p[0])
	}},
	0x07: command{"RLCA", 0, 4, func(c *Cpu) {
		c.a.set(c.rlc(c.a.Byte()))
	}},
	0x08: command{"LD (nn), SP", 2, 20, func(c *Cpu) {
		c.writeWord(BytesToWord(c.inst.p[1], c.inst.p[0]), c.sp)
	}},
	0x09: command{"", 0, 0, func(c *Cpu) {}},
	0x0A: command{"", 0, 0, func(c *Cpu) {}},
	0x0B: command{"DEC BC", 0, 8, func(c *Cpu) {
		c.b.setWord(c.b.Word() - 1)
	}},
	0x0C: command{"INC C", 0, 4, func(c *Cpu) {
		c.c.set(c.inc(c.c.Byte()))
	}},
	0x0D: command{"DEC C", 0, 4, func(c *Cpu) {
		c.c.set(c.dec(c.c.Byte()))
	}},
	0x0E: command{"LD C, #", 1, 8, func(c *Cpu) {
		c.c.set(c.inst.p[0])
	}},
	0x0F: command{"", 0, 0, func(c *Cpu) {}},
	0x10: command{"", 0, 0, func(c *Cpu) {}},
	0x11: command{"LD DE, nn", 2, 12, func(c *Cpu) {
		c.d.setWord(BytesToWord(c.inst.p[1], c.inst.p[0]))
	}},
	0x12: command{"LD (DE), A", 0, 8, func(c *Cpu) {
		c.writeByte(c.d.Word(), c.a.Byte())
	}},
	0x13: command{"INC DE", 0, 8, func(c *Cpu) {
		c.d.setWord(c.d.Word() + 1)
	}},
	0x14: command{"INC D", 0, 4, func(c *Cpu) {
		c.d.set(c.inc(c.d.Byte()))
	}},
	0x15: command{"DEC D", 0, 4, func(c *Cpu) {
		c.d.set(c.dec(c.d.Byte()))
	}},
	0x16: command{"LD D, #", 1, 8, func(c *Cpu) {
		c.d.set(c.inst.p[0])
	}},
	0x17: command{"RLA", 0, 4, func(c *Cpu) {
		c.a.set(c.rl(c.a.Byte()))
	}},
	0x18: command{"JR n", 1, 8, func(c *Cpu) {
		c.jr(int8(c.inst.p[0]))
	}},
	0x19: command{"", 0, 0, func(c *Cpu) {}},
	0x1A: command{"LD A, (DE)", 0, 8, func(c *Cpu) {
		c.a.set(c.readByte(c.d.Word()))
	}},
	0x1B: command{"", 0, 0, func(c *Cpu) {}},
	0x1C: command{"INC E", 0, 4, func(c *Cpu) {
		c.e.set(c.inc(c.e.Byte()))
	}},
	0x1D: command{"DEC E", 0, 4, func(c *Cpu) {
		c.e.set(c.dec(c.e.Byte()))
	}},
	0x1E: command{"LD E, #", 1, 8, func(c *Cpu) {
		c.e.set(c.inst.p[0])
	}},
	0x1F: command{"RRA", 0, 4, func(c *Cpu) {
		c.a.set(c.rr(c.a.Byte()))
	}},
	0x20: command{"JR NZ, *", 1, 8, func(c *Cpu) {
		c.jrNF(flagZ, int8(c.inst.p[0]))
	}},
	0x21: command{"LD HL, nn", 2, 12, func(c *Cpu) {
		c.h.setWord(BytesToWord(c.inst.p[1], c.inst.p[0]))
	}},
	0x22: command{"LDI (HL), A", 0, 8, func(c *Cpu) {
		c.writeByte(c.h.Word(), c.a.Byte())
		c.h.setWord(c.h.Word() + 1)
	}},
	0x23: command{"INC HL", 0, 8, func(c *Cpu) {
		c.h.setWord(c.h.Word() + 1)
	}},
	0x24: command{"INC H", 0, 4, func(c *Cpu) {
		c.h.set(c.inc(c.h.Byte()))
	}},
	0x25: command{"DEC H", 0, 4, func(c *Cpu) {
		c.h.set(c.dec(c.h.Byte()))
	}},
	0x26: command{"LD H, #", 1, 8, func(c *Cpu) {
		c.h.set(c.inst.p[0])
	}},
	0x27: command{"DAA", 0, 4, func(c *Cpu) {
		fmt.Println(c.str())
		panic("untested")
		a := c.a.Byte()
		if a&0x0F > 9 || c.f.getFlag(flagH) {
			a += 0x06
			c.f.setFlag(flagH)
		}
		if a > 0x9F || c.f.getFlag(flagC) {
			a += 0x60
			c.f.setFlag(flagC)
		}
	}},
	0x28: command{"JR Z, *", 1, 8, func(c *Cpu) {
		c.jrF(flagZ, int8(c.inst.p[0]))
	}},
	0x29: command{"", 0, 0, func(c *Cpu) {}},
	0x2A: command{"LDI A, (HL)", 0, 8, func(c *Cpu) {
		c.a.set(c.readByte(c.h.Word()))
		c.h.setWord(c.h.Word() + 1)
	}},
	0x2B: command{"", 0, 0, func(c *Cpu) {}},
	0x2C: command{"INC L", 0, 4, func(c *Cpu) {
		c.l.set(c.inc(c.l.Byte()))
	}},
	0x2D: command{"DEC L", 0, 4, func(c *Cpu) {
		c.l.set(c.dec(c.l.Byte()))
	}},
	0x2E: command{"LD L, #", 1, 8, func(c *Cpu) {
		c.l.set(c.inst.p[0])
	}},
	0x2F: command{"", 0, 0, func(c *Cpu) {}},
	0x30: command{"", 0, 0, func(c *Cpu) {}},
	0x31: command{"LD SP, nn", 2, 12, func(c *Cpu) {
		c.sp = register16(BytesToWord(c.inst.p[1], c.inst.p[0]))
	}},
	0x32: command{"LDD (HL), A", 0, 8, func(c *Cpu) {
		c.writeByte(c.h.Word(), c.a.Byte())
		c.h.setWord(c.h.Word() - 1)
	}},
	0x33: command{"", 0, 0, func(c *Cpu) {}},
	0x34: command{"INC (HL)", 0, 12, func(c *Cpu) {
		v := c.readByte(c.h.Word())
		v = c.inc(v)
		c.writeByte(c.h.Word(), v)
	}},
	0x35: command{"DEC (HL)", 0, 12, func(c *Cpu) {
		v := c.readByte(c.h.Word())
		v = c.dec(v)
		c.writeByte(c.h.Word(), v)
	}},
	0x36: command{"LD (HL), n", 1, 12, func(c *Cpu) {
		c.writeByte(c.h.Word(), c.inst.p[0])
	}},
	0x37: command{"", 0, 0, func(c *Cpu) {}},
	0x38: command{"", 0, 0, func(c *Cpu) {}},
	0x39: command{"", 0, 0, func(c *Cpu) {}},
	0x3A: command{"LDD A, (HL)", 0, 8, func(c *Cpu) {
		c.a.set(c.readByte(c.h.Word()))
		c.h.setWord(c.h.Word() - 1)
	}},
	0x3B: command{"", 0, 0, func(c *Cpu) {}},
	0x3C: command{"", 0, 0, func(c *Cpu) {}},
	0x3D: command{"DEC A", 0, 4, func(c *Cpu) {
		c.a.set(c.dec(c.a.Byte()))
	}},
	0x3E: command{"LD A, #", 1, 8, func(c *Cpu) {
		c.a.set(c.inst.p[0])
	}},
	0x3F: command{"", 0, 0, func(c *Cpu) {}},
	0x40: command{"LD B, B", 0, 4, func(c *Cpu) {
		c.b.set(c.b.Byte())
	}},
	0x41: command{"LD B, C", 0, 4, func(c *Cpu) {
		c.b.set(c.c.Byte())
	}},
	0x42: command{"LD B, D", 0, 4, func(c *Cpu) {
		c.b.set(c.d.Byte())
	}},
	0x43: command{"LD B, E", 0, 4, func(c *Cpu) {
		c.b.set(c.e.Byte())
	}},
	0x44: command{"LD B, H", 0, 4, func(c *Cpu) {
		c.b.set(c.h.Byte())
	}},
	0x45: command{"LD B, L", 0, 4, func(c *Cpu) {
		c.b.set(c.l.Byte())
	}},
	0x46: command{"LD B, (HL)", 0, 8, func(c *Cpu) {
		c.b.set(c.readByte(c.h.Word()))
	}},
	0x47: command{"LD B, A", 0, 4, func(c *Cpu) {
		c.b.set(c.a.Byte())
	}},
	0x48: command{"", 0, 0, func(c *Cpu) {}},
	0x49: command{"", 0, 0, func(c *Cpu) {}},
	0x4A: command{"", 0, 0, func(c *Cpu) {}},
	0x4B: command{"", 0, 0, func(c *Cpu) {}},
	0x4C: command{"", 0, 0, func(c *Cpu) {}},
	0x4D: command{"", 0, 0, func(c *Cpu) {}},
	0x4E: command{"", 0, 0, func(c *Cpu) {}},
	0x4F: command{"LD C, A", 0, 4, func(c *Cpu) {
		c.c.set(c.a.Byte())
	}},
	0x50: command{"", 0, 0, func(c *Cpu) {}},
	0x51: command{"", 0, 0, func(c *Cpu) {}},
	0x52: command{"", 0, 0, func(c *Cpu) {}},
	0x53: command{"", 0, 0, func(c *Cpu) {}},
	0x54: command{"", 0, 0, func(c *Cpu) {}},
	0x55: command{"", 0, 0, func(c *Cpu) {}},
	0x56: command{"", 0, 0, func(c *Cpu) {}},
	0x57: command{"LD D, A", 0, 4, func(c *Cpu) {
		c.d.set(c.a.Byte())
	}},
	0x58: command{"", 0, 0, func(c *Cpu) {}},
	0x59: command{"", 0, 0, func(c *Cpu) {}},
	0x5A: command{"", 0, 0, func(c *Cpu) {}},
	0x5B: command{"", 0, 0, func(c *Cpu) {}},
	0x5C: command{"", 0, 0, func(c *Cpu) {}},
	0x5D: command{"", 0, 0, func(c *Cpu) {}},
	0x5E: command{"", 0, 0, func(c *Cpu) {}},
	0x5F: command{"LD E, A", 0, 4, func(c *Cpu) {
		c.e.set(c.a.Byte())
	}},
	0x60: command{"", 0, 0, func(c *Cpu) {}},
	0x61: command{"", 0, 0, func(c *Cpu) {}},
	0x62: command{"", 0, 0, func(c *Cpu) {}},
	0x63: command{"", 0, 0, func(c *Cpu) {}},
	0x64: command{"", 0, 0, func(c *Cpu) {}},
	0x65: command{"", 0, 0, func(c *Cpu) {}},
	0x66: command{"", 0, 0, func(c *Cpu) {}},
	0x67: command{"LD H, A", 0, 4, func(c *Cpu) {
		c.h.set(c.a.Byte())
	}},
	0x68: command{"", 0, 0, func(c *Cpu) {}},
	0x69: command{"", 0, 0, func(c *Cpu) {}},
	0x6A: command{"", 0, 0, func(c *Cpu) {}},
	0x6B: command{"", 0, 0, func(c *Cpu) {}},
	0x6C: command{"", 0, 0, func(c *Cpu) {}},
	0x6D: command{"", 0, 0, func(c *Cpu) {}},
	0x6E: command{"", 0, 0, func(c *Cpu) {}},
	0x6F: command{"LD L, A", 0, 4, func(c *Cpu) {
		c.l.set(c.a.Byte())
	}},
	0x70: command{"", 0, 0, func(c *Cpu) {}},
	0x71: command{"", 0, 0, func(c *Cpu) {}},
	0x72: command{"", 0, 0, func(c *Cpu) {}},
	0x73: command{"LD (HL), E", 0, 8, func(c *Cpu) {
		c.writeByte(c.h.Word(), c.e.Byte())
	}},
	0x74: command{"", 0, 0, func(c *Cpu) {}},
	0x75: command{"", 0, 0, func(c *Cpu) {}},
	0x76: command{"", 0, 0, func(c *Cpu) {}},
	0x77: command{"LD (HL), A", 0, 8, func(c *Cpu) {
		c.writeByte(c.h.Word(), c.a.Byte())
	}},
	0x78: command{"LD A, B", 0, 4, func(c *Cpu) {
		c.a.set(c.b.Byte())
	}},
	0x79: command{"LD A, C", 0, 4, func(c *Cpu) {
		c.a.set(c.c.Byte())
	}},
	0x7A: command{"LD A, D", 0, 4, func(c *Cpu) {
		c.a.set(c.d.Byte())
	}},
	0x7B: command{"LD A, E", 0, 4, func(c *Cpu) {
		c.a.set(c.e.Byte())
	}},
	0x7C: command{"LD A, H", 0, 4, func(c *Cpu) {
		c.a.set(c.h.Byte())
	}},
	0x7D: command{"LD A, L", 0, 4, func(c *Cpu) {
		c.a.set(c.l.Byte())
	}},
	0x7E: command{"LD A, (HL)", 0, 8, func(c *Cpu) {
		c.a.set(c.readByte(c.h.Word()))
	}},
	0x7F: command{"LD A, A", 0, 4, func(c *Cpu) {
		c.a.set(c.a.Byte())
	}},
	0x80: command{"ADD A, B", 0, 4, func(c *Cpu) {
		c.a.set(c.add(c.a.Byte(), c.b.Byte()))
	}},
	0x81: command{"ADD A, C", 0, 4, func(c *Cpu) {
		c.a.set(c.add(c.a.Byte(), c.c.Byte()))
	}},
	0x82: command{"ADD A, D", 0, 4, func(c *Cpu) {
		c.a.set(c.add(c.a.Byte(), c.d.Byte()))
	}},
	0x83: command{"ADD A, E", 0, 4, func(c *Cpu) {
		c.a.set(c.add(c.a.Byte(), c.e.Byte()))
	}},
	0x84: command{"ADD A, H", 0, 4, func(c *Cpu) {
		c.a.set(c.add(c.a.Byte(), c.h.Byte()))
	}},
	0x85: command{"ADD A, L", 0, 4, func(c *Cpu) {
		c.a.set(c.add(c.a.Byte(), c.l.Byte()))
	}},
	0x86: command{"ADD A, (HL)", 0, 8, func(c *Cpu) {
		c.a.set(c.add(c.a.Byte(), c.readByte(c.h.Word())))
	}},
	0x87: command{"ADD A, A", 0, 4, func(c *Cpu) {
		c.a.set(c.add(c.a.Byte(), c.a.Byte()))
	}},
	0x88: command{"ADC A, B", 0, 4, func(c *Cpu) {
		c.a.set(c.adc(c.a.Byte(), c.b.Byte()))
	}},
	0x89: command{"ADC A, C", 0, 4, func(c *Cpu) {
		c.a.set(c.adc(c.a.Byte(), c.c.Byte()))
	}},
	0x8A: command{"ADC A, D", 0, 4, func(c *Cpu) {
		c.a.set(c.adc(c.a.Byte(), c.d.Byte()))
	}},
	0x8B: command{"ADC A, E", 0, 4, func(c *Cpu) {
		c.a.set(c.adc(c.a.Byte(), c.e.Byte()))
	}},
	0x8C: command{"ADC A, H", 0, 4, func(c *Cpu) {
		c.a.set(c.adc(c.a.Byte(), c.h.Byte()))
	}},
	0x8D: command{"ADC A, L", 0, 4, func(c *Cpu) {
		c.a.set(c.adc(c.a.Byte(), c.l.Byte()))
	}},
	0x8E: command{"ADC A, (HL)", 0, 8, func(c *Cpu) {
		c.a.set(c.adc(c.a.Byte(), c.readByte(c.h.Word())))
	}},
	0x8F: command{"ADC A, A", 0, 4, func(c *Cpu) {
		c.a.set(c.adc(c.a.Byte(), c.a.Byte()))
	}},
	0x90: command{"SUB B", 0, 4, func(c *Cpu) {
		c.a.set(c.sub(c.a.Byte(), c.b.Byte()))
	}},
	0x91: command{"SUB C", 0, 4, func(c *Cpu) {
		c.a.set(c.sub(c.a.Byte(), c.c.Byte()))
	}},
	0x92: command{"SUB D", 0, 4, func(c *Cpu) {
		c.a.set(c.sub(c.a.Byte(), c.d.Byte()))
	}},
	0x93: command{"SUB E", 0, 4, func(c *Cpu) {
		c.a.set(c.sub(c.a.Byte(), c.e.Byte()))
	}},
	0x94: command{"SUB H", 0, 4, func(c *Cpu) {
		c.a.set(c.sub(c.a.Byte(), c.h.Byte()))
	}},
	0x95: command{"SUB L", 0, 4, func(c *Cpu) {
		c.a.set(c.sub(c.a.Byte(), c.l.Byte()))
	}},
	0x96: command{"SUB (HL)", 0, 8, func(c *Cpu) {
		v := c.readByte(c.h.Word())
		c.a.set(c.sub(c.a.Byte(), v))
	}},
	0x97: command{"SUB A", 0, 4, func(c *Cpu) {
		c.a.set(c.sub(c.a.Byte(), c.a.Byte()))
	}},
	0x98: command{"", 0, 0, func(c *Cpu) {}},
	0x99: command{"", 0, 0, func(c *Cpu) {}},
	0x9A: command{"", 0, 0, func(c *Cpu) {}},
	0x9B: command{"", 0, 0, func(c *Cpu) {}},
	0x9C: command{"", 0, 0, func(c *Cpu) {}},
	0x9D: command{"", 0, 0, func(c *Cpu) {}},
	0x9E: command{"", 0, 0, func(c *Cpu) {}},
	0x9F: command{"", 0, 0, func(c *Cpu) {}},
	0xA0: command{"", 0, 0, func(c *Cpu) {}},
	0xA1: command{"", 0, 0, func(c *Cpu) {}},
	0xA2: command{"", 0, 0, func(c *Cpu) {}},
	0xA3: command{"", 0, 0, func(c *Cpu) {}},
	0xA4: command{"AND H", 0, 4, func(c *Cpu) {
		c.a.set(c.and(c.a.Byte(), c.h.Byte()))
	}},
	0xA5: command{"", 0, 0, func(c *Cpu) {}},
	0xA6: command{"", 0, 0, func(c *Cpu) {}},
	0xA7: command{"", 0, 0, func(c *Cpu) {}},
	0xA8: command{"XOR B", 0, 4, func(c *Cpu) {
		c.a.set(c.xor(c.a.Byte(), c.b.Byte()))
	}},
	0xA9: command{"XOR C", 0, 4, func(c *Cpu) {
		c.a.set(c.xor(c.a.Byte(), c.c.Byte()))
	}},
	0xAA: command{"XOR D", 0, 4, func(c *Cpu) {
		c.a.set(c.xor(c.a.Byte(), c.d.Byte()))
	}},
	0xAB: command{"XOR E", 0, 4, func(c *Cpu) {
		c.a.set(c.xor(c.a.Byte(), c.e.Byte()))
	}},
	0xAC: command{"XOR H", 0, 4, func(c *Cpu) {
		c.a.set(c.xor(c.a.Byte(), c.h.Byte()))
	}},
	0xAD: command{"XOR L", 0, 4, func(c *Cpu) {
		c.a.set(c.xor(c.a.Byte(), c.l.Byte()))
	}},
	0xAE: command{"XOR (HL)", 0, 8, func(c *Cpu) {
		c.a.set(c.xor(c.a.Byte(), c.readByte(c.h.Word())))
	}},
	0xAF: command{"XOR A", 0, 4, func(c *Cpu) {
		c.a.set(c.xor(c.a.Byte(), c.a.Byte()))
	}},
	0xB0: command{"OR B", 0, 4, func(c *Cpu) {
		c.a.set(c.or(c.a.Byte(), c.b.Byte()))
	}},
	0xB1: command{"OR C", 0, 4, func(c *Cpu) {
		c.a.set(c.or(c.a.Byte(), c.c.Byte()))
	}},
	0xB2: command{"OR D", 0, 4, func(c *Cpu) {
		c.a.set(c.or(c.a.Byte(), c.d.Byte()))
	}},
	0xB3: command{"OR E", 0, 4, func(c *Cpu) {
		c.a.set(c.or(c.a.Byte(), c.e.Byte()))
	}},
	0xB4: command{"OR H", 0, 4, func(c *Cpu) {
		c.a.set(c.or(c.a.Byte(), c.h.Byte()))
	}},
	0xB5: command{"OR L", 0, 4, func(c *Cpu) {
		c.a.set(c.or(c.a.Byte(), c.l.Byte()))
	}},
	0xB6: command{"OR (HL)", 0, 8, func(c *Cpu) {
		c.a.set(c.or(c.a.Byte(), c.readByte(c.h.Word())))
	}},
	0xB7: command{"", 0, 0, func(c *Cpu) {}},

	0xB8: command{"CP B", 0, 4, func(c *Cpu) {
		c.sub(c.a.Byte(), c.b.Byte())
	}},
	0xB9: command{"CP C", 0, 4, func(c *Cpu) {
		c.sub(c.a.Byte(), c.c.Byte())
	}},
	0xBA: command{"CP D", 0, 4, func(c *Cpu) {
		c.sub(c.a.Byte(), c.d.Byte())
	}},
	0xBB: command{"CP E", 0, 4, func(c *Cpu) {
		c.sub(c.a.Byte(), c.e.Byte())
	}},
	0xBC: command{"CP H", 0, 4, func(c *Cpu) {
		c.sub(c.a.Byte(), c.h.Byte())
	}},
	0xBD: command{"CP L", 0, 4, func(c *Cpu) {
		c.sub(c.a.Byte(), c.l.Byte())
	}},
	0xBE: command{"CP (HL)", 0, 8, func(c *Cpu) {
		v := c.readByte(c.h.Word())
		c.sub(c.a.Byte(), v)
	}},
	0xBF: command{"CP A", 0, 4, func(c *Cpu) {
		c.sub(c.a.Byte(), c.a.Byte())
	}},
	0xC0: command{"", 0, 0, func(c *Cpu) {}},
	0xC1: command{"POP BC", 0, 12, func(c *Cpu) {
		c.b.setWord(c.pop())
	}},
	0xC2: command{"", 0, 0, func(c *Cpu) {}},
	0xC3: command{"JP nn", 2, 12, func(c *Cpu) {
		c.jp(BytesToWord(c.inst.p[1], c.inst.p[0]))
	}},
	0xC4: command{"", 0, 0, func(c *Cpu) {}},
	0xC5: command{"PUSH BC", 0, 16, func(c *Cpu) {
		c.push(c.b.Word())
	}},
	0xC6: command{"", 0, 0, func(c *Cpu) {}},
	0xC7: command{"", 0, 0, func(c *Cpu) {}},
	0xC8: command{"", 0, 0, func(c *Cpu) {}},
	0xC9: command{"RET", 0, 8, func(c *Cpu) {
		c.jp(c.pop())
	}},
	0xCA: command{"", 0, 0, func(c *Cpu) {}},
	0xCB01: command{"RLC C", 0, 8, func(c *Cpu) {
		c.c.set(c.rlc(c.c.Byte()))
	}},
	0xCB11: command{"RL C", 0, 8, func(c *Cpu) {
		c.c.set(c.rl(c.c.Byte()))
	}},
	0xCB7C: command{"BIT 7, H", 0, 8, func(c *Cpu) {
		c.bit(7, c.h.Byte())
	}},
	0xCC: command{"CALL Z, nn", 2, 12, func(c *Cpu) {
		c.callF(flagZ, BytesToWord(c.inst.p[1], c.inst.p[0]))
	}},
	0xCD: command{"CALL nn", 2, 12, func(c *Cpu) {
		c.call(BytesToWord(c.inst.p[1], c.inst.p[0]))
	}},
	0xCE: command{"", 0, 0, func(c *Cpu) {}},
	0xCF: command{"", 0, 0, func(c *Cpu) {}},
	0xD0: command{"", 0, 0, func(c *Cpu) {}},
	0xD1: command{"", 0, 0, func(c *Cpu) {}},
	0xD2: command{"", 0, 0, func(c *Cpu) {}},
	0xD3: command{"", 0, 0, func(c *Cpu) {}},
	0xD4: command{"", 0, 0, func(c *Cpu) {}},
	0xD5: command{"", 0, 0, func(c *Cpu) {}},
	0xD6: command{"", 0, 0, func(c *Cpu) {}},
	0xD7: command{"", 0, 0, func(c *Cpu) {}},
	0xD8: command{"", 0, 0, func(c *Cpu) {}},
	0xD9: command{"", 0, 0, func(c *Cpu) {}},
	0xDA: command{"", 0, 0, func(c *Cpu) {}},
	0xDB: command{"", 0, 0, func(c *Cpu) {}},
	0xDC: command{"", 0, 0, func(c *Cpu) {}},
	0xDE: command{"", 0, 0, func(c *Cpu) {}},
	0xDF: command{"", 0, 0, func(c *Cpu) {}},
	0xE0: command{"LDH (n), A", 1, 12, func(c *Cpu) {
		c.writeByte(Word(0xFF00+uint16(c.inst.p[0])), c.a.Byte())
	}},
	0xE1: command{"", 0, 0, func(*Cpu) {}},
	0xE2: command{"LD (C), A", 0, 8, func(c *Cpu) {
		c.writeByte(Word(0xFF00+uint16(c.c.Byte())), c.a.Byte())
	}},
	0xE3: command{"", 0, 0, func(c *Cpu) {}},
	0xE4: command{"", 0, 0, func(c *Cpu) {}},
	0xE5: command{"", 0, 0, func(c *Cpu) {}},
	0xE6: command{"", 0, 0, func(c *Cpu) {}},
	0xE7: command{"", 0, 0, func(c *Cpu) {}},
	0xE8: command{"", 0, 0, func(c *Cpu) {}},
	0xE9: command{"", 0, 0, func(c *Cpu) {}},
	0xEA: command{"LD (nn), A", 2, 16, func(c *Cpu) {
		c.writeByte(BytesToWord(c.inst.p[1], c.inst.p[0]), c.a.Byte())
	}},
	0xEB: command{"", 0, 0, func(c *Cpu) {}},
	0xEC: command{"", 0, 0, func(c *Cpu) {}},
	0xED: command{"", 0, 0, func(c *Cpu) {}},
	0xEE: command{"", 0, 0, func(c *Cpu) {}},
	0xEF: command{"", 0, 0, func(c *Cpu) {}},
	0xF0: command{"LDH A, (n)", 1, 12, func(c *Cpu) {
		c.a.set(c.readByte(Word(0xFF00 + uint16(c.inst.p[0]))))
	}},
	0xF1: command{"", 0, 0, func(c *Cpu) {}},
	0xF2: command{"LD A, (C)", 0, 8, func(c *Cpu) {
		c.a.set(c.readByte(Word(0xFF00 + uint16(c.c.Byte()))))
	}},
	0xF3: command{"DI", 0, 4, func(c *Cpu) {
		c.ime = Bit(0)
	}},
	0xF4: command{"", 0, 0, func(c *Cpu) {}},
	0xF5: command{"", 0, 0, func(c *Cpu) {}},
	0xF6: command{"", 0, 0, func(c *Cpu) {}},
	0xF7: command{"", 0, 0, func(c *Cpu) {}},
	0xF8: command{"LDHL SP, n", 1, 12, func(c *Cpu) {
		fmt.Println(c.str())
		panic("untested")
		c.h.setWord(c.addWordR(c.sp, c.inst.p[0]))
		c.f.resetFlag(flagZ)
		c.f.resetFlag(flagN)
	}},
	0xF9: command{"", 0, 0, func(c *Cpu) {}},
	0xFA: command{"LD A, (nn)", 2, 16, func(c *Cpu) {
		nn := BytesToWord(c.inst.p[1], c.inst.p[0])
		c.a.set(c.readByte(nn))
	}},
	0xFB: command{"", 0, 0, func(c *Cpu) {}},
	0xFC: command{"", 0, 0, func(c *Cpu) {}},
	0xFD: command{"", 0, 0, func(c *Cpu) {}},
	0xFE: command{"CP #", 1, 8, func(c *Cpu) {
		c.sub(c.a.Byte(), c.inst.p[0])
	}},
	0xFF: command{"", 0, 0, func(*Cpu) {}},
}
