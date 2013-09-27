package main

import (
	"fmt"
	"time"
)

// 1 machine cycle = 4 clock cycles
// machine cycles: 1.05MHz nop: 1 cycle
// clock cycles: 4.19MHz nop: 4 cycles
type cpu struct {
	a  uint8
	b  uint8
	c  uint8
	d  uint8
	e  uint8
	f  uint8 // lower 4 bits always read cero
	h  uint8
	l  uint8
	sp uint16
	pc uint16

	rom []uint8
	ram []uint8

	period time.Duration

	clock  <-chan time.Time
	cycles uint8
}

const (
	flagZReset uint8 = 0x70
	flagNReset       = 0xB0
	flagHReset       = 0xD0
	flagCReset       = 0xE0

	flagZSet = 0x80
	flagNSet = 0x40
	flagHSet = 0x20
	flagCSet = 0x10
)

func newCpu(rom []uint8) *cpu {
	romFull := make([]uint8, 256)
	copy(romFull, rom)
	ram := make([]uint8, 65536)
	hz := 1.05e6
	period := time.Duration(1e9 / hz)
	clock := time.Tick(period)

	return &cpu{sp: 0xFFFE, rom: romFull, ram: ram, period: period, clock: clock}
}

func (c *cpu) String() string {
	return fmt.Sprintf("a:%v b:%v c:%v d:%v e:%v f:%v h:%v l:%v sp:%v pc:%v",
		c.a, c.b, c.c, c.d, c.e, c.f, c.h, c.l, c.sp, c.pc)
}

func (c *cpu) readPc() uint8 {
	r := c.readByte(uint8(c.pc>>8), uint8(c.pc&0xFF))
	c.pc++
	return r
}

func (c *cpu) readByte(ms, ls uint8) uint8 {
	<-c.clock
	c.cycles++
	addr := uint16(ms)<<8 + uint16(ls)
	if addr < 256 {
		return c.rom[addr]
	}
	return c.ram[addr]
}

func (c *cpu) writeByte(ms, ls, b uint8) {
	<-c.clock
	c.cycles++
	addr := uint16(ms)<<8 + uint16(ls)
	c.ram[addr] = b
}

func (c *cpu) subWord(a uint16, b uint16) uint16 {
	//TODO: set flags
	return a - b
}

func (c *cpu) addSigned(a uint16, b int8) uint16 {
	if b < 0 {
		return c.subWord(a, uint16(-b))
	}
	//TODO: set flags
	return a + uint16(b)
}

func (c *cpu) add(a, b uint8) uint8 {
	r := a + b
	c.f = 0
	if r == 0 {
		c.f |= flagZSet
	}
	if a&0x0F+b&0x0F > 0x0F {
		c.f |= flagHSet
	}
	if uint16(a)+uint16(b) > 0xFF {
		c.f |= flagCSet
	}
	return r
}

func (c *cpu) addC(a, b uint8) uint8 {
	carry := uint8(0)
	if c.f&flagCSet > 0 {
		carry = 1
	}
	r := a + b + carry
	c.f = 0
	if r == 0 {
		c.f |= flagZSet
	}
	if a&0x0F+b&0x0F+carry > 0x0F {
		c.f |= flagHSet
	}
	if uint16(a)+uint16(b)+uint16(carry) > 0xFF {
		c.f |= flagCSet
	}
	return r
}

func (c *cpu) sub(a, b uint8) uint8 {
	r := a - b
	c.f = 0
	if r == 0 {
		c.f |= flagZSet
	}
	c.f |= flagNSet
	if a&0x0F >= b&0x0F {
		c.f |= flagHSet
	}
	if a >= b {
		c.f |= flagCSet
	}
	return r
}

func (c *cpu) subC(a, b uint8) uint8 {
	carry := uint8(0)
	if c.f&flagCSet > 0 {
		carry = 1
	}
	r := a - b - carry
	c.f = 0
	if r == 0 {
		c.f |= flagZSet
	}
	c.f |= flagNSet
	if a&0x0F >= (b&0x0F + carry) {
		c.f |= flagHSet
	}
	if a >= (b + carry) {
		c.f |= flagCSet
	}
	return r
}

func (c *cpu) and(a, b uint8) uint8 {
	r := a & b
	c.f = 0
	if r == 0 {
		c.f |= flagZSet
	}
	c.f |= flagHSet
	return r
}

func (c *cpu) or(a, b uint8) uint8 {
	r := a | b
	c.f = 0
	if r == 0 {
		c.f |= flagZSet
	}
	return r
}

func (c *cpu) push(b uint8) {
	c.writeByte(uint8(c.sp>>8), uint8(c.sp&0xFF), b)
	c.sp--
}

func (c *cpu) pop() uint8 {
	c.sp++
	r := c.readByte(uint8(c.sp>>8), uint8(c.sp&0xFF))
	return r
}

func decrement(ms, ls uint8) (uint8, uint8) {
	addr := uint16(ms)<<8 + uint16(ls) - 1
	return uint8(addr >> 8), uint8(addr & 0xFF)
}

func increment(ms, ls uint8) (uint8, uint8) {
	addr := uint16(ms)<<8 + uint16(ls) + 1
	return uint8(addr >> 8), uint8(addr & 0xFF)
}

func main() {
	c := newCpu([]uint8{0x06, 55, 0x0A, 0x7e, 0x22, 0x03, 0xF5, 0x0A, 0xF1, 0x80, 0xFF})

	// main loop
	startTime := time.Now()
	for {
		opcode := c.readPc()
		switch opcode {
		case 0x00: // NOP
		case 0x01: // LD BC, nn
			c.c = c.readPc()
			c.b = c.readPc()
		case 0x02: // LD (BC), A
			c.writeByte(c.b, c.c, c.a)
		case 0x03: // INC BC
			c.b, c.c = increment(c.b, c.c)
			<-c.clock
			c.cycles++
		case 0x06: // LD B, n
			c.b = c.readPc()
		case 0x08: // LD (nn), SP
			l := c.readPc()
			h := c.readPc()
			c.writeByte(h, l, uint8(c.sp&0xFF))
			h, l = increment(h, l)
			c.writeByte(h, l, uint8(c.sp>>8))
		case 0x0A: // LD A, (BC)
			c.a = c.readByte(c.b, c.c)
		case 0x0E: // LD C, n
			c.c = c.readPc()
		case 0x11: // LD DE, nn
			c.e = c.readPc()
			c.d = c.readPc()
		case 0x12: // LD (DE), A
			c.writeByte(c.d, c.e, c.a)
		case 0x16: // LD D, n
			c.d = c.readPc()
		case 0x1A: // LD A, (DE)
			c.a = c.readByte(c.d, c.e)
		case 0x1E: // LD E, n
			c.e = c.readPc()
		case 0x21: // LD HL, nn
			c.l = c.readPc()
			c.h = c.readPc()
		case 0x22: // LDI (HL), A
			c.writeByte(c.h, c.l, c.a)
			c.h, c.l = increment(c.h, c.l)
		case 0x26: // LD H, n
			c.h = c.readPc()
		case 0x2A: // LDI A, (HL)
			c.a = c.readByte(c.h, c.l)
			c.h, c.l = increment(c.h, c.l)
		case 0x2E: // LD L, n
			c.l = c.readPc()
		case 0x31: // LD SP, nn
			l := c.readPc()
			h := c.readPc()
			c.sp = uint16(h)<<8 + uint16(l)
		case 0x32: // LDD (HL), A
			c.writeByte(c.h, c.l, c.a)
			c.h, c.l = decrement(c.h, c.l)
		case 0x36: // LD (HL), n
			c.writeByte(c.h, c.l, c.readPc())
		case 0x3A: // LDD A, (HL)
			c.a = c.readByte(c.h, c.l)
			c.h, c.l = decrement(c.h, c.l)
		case 0x3E: // LD A, n
			c.a = c.readPc()
		case 0x40: // LD B, B
			c.b = c.b
		case 0x41: // LD B, C
			c.b = c.c
		case 0x42: // LD B, D
			c.b = c.d
		case 0x43: // LD B, E
			c.b = c.e
		case 0x44: // LD B, H
			c.b = c.h
		case 0x45: // LD B, L
			c.b = c.l
		case 0x46: // LD B, (HL)
			c.b = c.readByte(c.h, c.l)
		case 0x47: // LD B, A
			c.b = c.a
		case 0x48: // LD C, B
			c.c = c.b
		case 0x49: // LD C, C
			c.c = c.c
		case 0x4A: // LD C, D
			c.c = c.d
		case 0x4B: // LD C, E
			c.c = c.e
		case 0x4C: // LD C, H
			c.c = c.h
		case 0x4D: // LD C, L
			c.c = c.l
		case 0x4E: // LD C, (HL)
			c.c = c.readByte(c.h, c.l)
		case 0x4F: // LD C, A
			c.c = c.a
		case 0x50: // LD D, B
			c.d = c.b
		case 0x51: // LD D, C
			c.d = c.c
		case 0x52: // LD D, D
			c.d = c.d
		case 0x53: // LD D, E
			c.d = c.e
		case 0x54: // LD D, H
			c.d = c.h
		case 0x55: // LD D, L
			c.d = c.l
		case 0x56: // LD D, (HL)
			c.d = c.readByte(c.h, c.l)
		case 0x57: // LD D, A
			c.d = c.a
		case 0x58: // LD E, B
			c.e = c.b
		case 0x59: // LD E, C
			c.e = c.c
		case 0x5A: // LD E, D
			c.e = c.d
		case 0x5B: // LD E, E
			c.e = c.e
		case 0x5C: // LD E, H
			c.e = c.h
		case 0x5D: // LD E, L
			c.e = c.l
		case 0x5E: // LD E, (HL)
			c.e = c.readByte(c.h, c.l)
		case 0x5F: // LD E, A
			c.e = c.a
		case 0x60: // LD H, B
			c.h = c.b
		case 0x61: // LD H, C
			c.h = c.c
		case 0x62: // LD H, D
			c.h = c.d
		case 0x63: // LD H, E
			c.h = c.e
		case 0x64: // LD H, H
			c.h = c.h
		case 0x65: // LD H, L
			c.h = c.l
		case 0x66: // LD H, (HL)
			c.h = c.readByte(c.h, c.l)
		case 0x67: // LD H, A
			c.h = c.a
		case 0x68: // LD L, B
			c.l = c.b
		case 0x69: // LD L, C
			c.l = c.c
		case 0x6A: // LD L, D
			c.l = c.d
		case 0x6B: // LD L, E
			c.l = c.e
		case 0x6C: // LD L, H
			c.l = c.h
		case 0x6D: // LD L, L
			c.l = c.l
		case 0x6E: // LD L, (HL)
			c.l = c.readByte(c.h, c.l)
		case 0x6F: // LD L, A
			c.l = c.a
		case 0x70: // LD (HL), B
			c.writeByte(c.h, c.l, c.b)
		case 0x71: // LD (HL), C
			c.writeByte(c.h, c.l, c.c)
		case 0x72: // LD (HL), D
			c.writeByte(c.h, c.l, c.d)
		case 0x73: // LD (HL), E
			c.writeByte(c.h, c.l, c.e)
		case 0x74: // LD (HL), H
			c.writeByte(c.h, c.l, c.h)
		case 0x75: // LD (HL), l
			c.writeByte(c.h, c.l, c.l)
		case 0x77: // LD (HL), A
			c.writeByte(c.h, c.l, c.a)
		case 0x78: // LD A, B
			c.a = c.b
		case 0x79: // LD A, C
			c.a = c.c
		case 0x7A: // LD A, D
			c.a = c.d
		case 0x7B: // LD A, E
			c.a = c.e
		case 0x7C: // LD A, H
			c.a = c.h
		case 0x7D: // LD A, L
			c.a = c.l
		case 0x7E: // LD A, (HL)
			c.a = c.readByte(c.h, c.l)
		case 0x7F: // LD A, A
			c.a = c.a
		case 0x80: // ADD A, B
			c.a = c.add(c.a, c.b)
		case 0x81: // ADD A, C
			c.a = c.add(c.a, c.c)
		case 0x82: // ADD A, D
			c.a = c.add(c.a, c.d)
		case 0x83: // ADD A, E
			c.a = c.add(c.a, c.e)
		case 0x84: // ADD A, H
			c.a = c.add(c.a, c.h)
		case 0x85: // ADD A, L
			c.a = c.add(c.a, c.l)
		case 0x86: // ADD A, (HL)
			c.a = c.add(c.a, c.readByte(c.h, c.l))
		case 0x87: // ADD A, A
			c.a = c.add(c.a, c.a)
		case 0x88: // ADC A, B
			c.a = c.addC(c.a, c.b)
		case 0x89: // ADC A, C
			c.a = c.addC(c.a, c.c)
		case 0x8A: // ADC A, D
			c.a = c.addC(c.a, c.d)
		case 0x8B: // ADC A, E
			c.a = c.addC(c.a, c.e)
		case 0x8C: // ADC A, H
			c.a = c.addC(c.a, c.h)
		case 0x8D: // ADC A, L
			c.a = c.addC(c.a, c.l)
		case 0x8E: // ADC A, (HL)
			c.a = c.addC(c.a, c.readByte(c.h, c.l))
		case 0x8F: // ADC A, A
			c.a = c.addC(c.a, c.a)
		case 0x90: // SUB A, B
			c.a = c.sub(c.a, c.b)
		case 0x91: // SUB A, C
			c.a = c.sub(c.a, c.c)
		case 0x92: // SUB A, D
			c.a = c.sub(c.a, c.d)
		case 0x93: // SUB A, E
			c.a = c.sub(c.a, c.e)
		case 0x94: // SUB A, H
			c.a = c.sub(c.a, c.h)
		case 0x95: // SUB A, L
			c.a = c.sub(c.a, c.l)
		case 0x96: // SUB A, (HL)
			c.a = c.sub(c.a, c.readByte(c.h, c.l))
		case 0x97: // SUB A, A
			c.a = c.sub(c.a, c.a)
		case 0x98: // SBC A, B
			c.a = c.subC(c.a, c.b)
		case 0x99: // SBC A, C
			c.a = c.subC(c.a, c.c)
		case 0x9A: // SBC A, D
			c.a = c.subC(c.a, c.d)
		case 0x9B: // SBC A, E
			c.a = c.subC(c.a, c.e)
		case 0x9C: // SBC A, H
			c.a = c.subC(c.a, c.h)
		case 0x9D: // SBC A, L
			c.a = c.subC(c.a, c.l)
		case 0x9E: // SBC A, (HL)
			c.a = c.subC(c.a, c.readByte(c.h, c.l))
		case 0x9F: // SBC A, A
			c.a = c.subC(c.a, c.a)
		case 0xA0: // AND B
			c.a = c.and(c.a, c.b)
		case 0xA1: // AND C
			c.a = c.and(c.a, c.c)
		case 0xA2: // AND D
			c.a = c.and(c.a, c.d)
		case 0xA3: // AND E
			c.a = c.and(c.a, c.e)
		case 0xA4: // AND H
			c.a = c.and(c.a, c.h)
		case 0xA5: // AND L
			c.a = c.and(c.a, c.l)
		case 0xA6: // AND (HL)
			c.a = c.and(c.a, c.readByte(c.h, c.l))
		case 0xA7: // AND A
			c.a = c.and(c.a, c.a)
		case 0xB0: // OR B
			c.a = c.or(c.a, c.b)
		case 0xB1: // OR C
			c.a = c.or(c.a, c.c)
		case 0xB2: // OR D
			c.a = c.or(c.a, c.d)
		case 0xB3: // OR E
			c.a = c.or(c.a, c.e)
		case 0xB4: // OR H
			c.a = c.or(c.a, c.h)
		case 0xB5: // OR L
			c.a = c.or(c.a, c.l)
		case 0xB6: // OR (HL)
			c.a = c.or(c.a, c.readByte(c.h, c.l))
		case 0xB7: // OR A
			c.a = c.or(c.a, c.a)
		case 0xC1: // POP BC
			c.c = c.pop()
			c.b = c.pop()
		case 0xC5: // PUSH BC
			c.push(c.b)
			c.push(c.c)
			<-c.clock
			c.cycles++
		case 0xC6: // ADD A, #
			c.a = c.add(c.a, c.readPc())
		case 0xCE: // ADC A, #
			c.a = c.addC(c.a, c.readPc())
		case 0xD1: // POP DE
			c.e = c.pop()
			c.d = c.pop()
		case 0xD5: // PUSH DE
			c.push(c.d)
			c.push(c.e)
			<-c.clock
			c.cycles++
		case 0xD6: // SUB A, #
			c.a = c.sub(c.a, c.readPc())
		case 0xDE: // SBC A, #
			c.a = c.subC(c.a, c.readPc())
		case 0xE0: // LDH (n), A
			n := c.readPc()
			c.writeByte(0xFF, n, c.a)
		case 0xE1: // POP HL
			c.l = c.pop()
			c.h = c.pop()
		case 0xE2: // LD (C), A
			c.writeByte(0xFF, c.c, c.a)
		case 0xE5: // PUSH HL
			c.push(c.h)
			c.push(c.l)
			<-c.clock
			c.cycles++
		case 0xE6: // AND #
			c.a = c.and(c.a, c.readPc())
		case 0xEA: // LD (nn), A
			l := c.readPc()
			h := c.readPc()
			c.writeByte(h, l, c.a)
		case 0xF0: // LDH A, (n)
			n := c.readPc()
			c.a = c.readByte(0xFF, n)
		case 0xF1: // POP AF
			c.f = c.pop()
			c.a = c.pop()
		case 0xF2: // LD A, (C)
			c.a = c.readByte(0xFF, c.c)
		case 0xF5: // PUSH AF
			c.push(c.a)
			c.push(c.f)
			<-c.clock
			c.cycles++
		case 0xF8: // LDHL SP, n
			n := int8(c.readPc())
			hl := c.addSigned(c.sp, n)
			c.h = uint8(hl >> 8)
			c.l = uint8(hl & 0xFF)
			c.f &= flagZReset
			c.f &= flagNReset
			<-c.clock
			c.cycles++
		case 0xF9: // LD SP, HL
			c.sp = uint16(c.h)<<8 + uint16(c.l)
			<-c.clock
			c.cycles++
		case 0xFA: // LD A, (nn)
			l := c.readPc()
			h := c.readPc()
			c.a = c.readByte(h, l)
		default:
			panic(fmt.Sprintf("unknown opcode 0x%X", opcode))
		}
		period := time.Since(startTime)
		startTime = time.Now()
		//mhz := 1e3 * float64(c.cycles) / float64(period)

		//if c.pc == 0 {
		fmt.Println(c, period)
		//}

		c.cycles = 0
	}
}
