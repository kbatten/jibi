package main

import (
	"fmt"
	"time"
)

type cpu struct {
	// registers
	a  register8
	b  register8
	c  register8
	d  register8
	e  register8
	f  register8 // 8 bits, but lower 4 bits always read zero
	h  register8
	l  register8
	sp register16
	pc register16

	// clocks
	mClock  <-chan time.Time
	mTicker *time.Ticker
	m       uint8 // machine cycles
	t       uint8 // clock cycles

	// current instruction buffer
	inst     instruction
	commands []command

	// extra state
	// TODO: find a way to remove
	di int // disable interrupts counter
	ei int // enable interrutps counter

	// connections
	mc  memoryController // read/write bytes and words
	res connection       // reset cpu on read
}

const (
	flagZ = 0x80
	flagN = 0x40
	flagH = 0x20
	flagC = 0x10
)

func newCpu(mc memoryController, reset connection) *cpu {
	// use internal clock
	// 1 machine cycle = 4 clock cycles
	// machine cycles: 1.05MHz nop: 1 cycle
	// clock cycles: 4.19MHz nop: 4 cycles
	hz := 4.194304 * 1e6 / 4.0 // 4.19MHz clock to 1.05 machine cycles
	period := time.Duration(1e9 / hz)
	ticker := time.NewTicker(period)
	clock := ticker.C

	f := newRegister8(0xF0, nil)
	a := newRegister8(0, &f)
	c := newRegister8(0, nil)
	b := newRegister8(0, &c)
	e := newRegister8(0, nil)
	d := newRegister8(0, &e)
	l := newRegister8(0, nil)
	h := newRegister8(0, &l)

	return &cpu{a: a, b: b, c: c, d: d, e: e, f: f, l: l, h: h,
		sp: 0xFFFE, mTicker: ticker, mClock: clock, res:reset, mc:mc}
}

func (c *cpu) String() string {
	return fmt.Sprintf(`%v
    a:%v b:%v c:%v d:%v e:%v f:%v h:%v l:%v
    af:0x%04X bc:0x%04X de:0x%04X hl:0x%04X sp:%v pc:%v`,
		c.inst, c.a, c.b, c.c, c.d, c.e, c.f, c.h, c.l,
		c.a.getWord(), c.b.getWord(), c.d.getWord(), c.h.getWord(), c.sp, c.pc)
}

func (c *cpu) reset() {
	c.a.set(0)
	c.b.set(0)
	c.c.set(0)
	c.d.set(0)
	c.e.set(0)
	c.f.set(0)
	c.h.set(0)
	c.l.set(0)
	c.sp = 0xFFFE
	c.pc = 0
	c.m = 0
	c.t = 0
	c.di = 0
	c.ei = 0
}

/*
func (c *cpu) readPc() uint8 {
	<-c.clock
	c.cycles++
	r := c.rom[c.pc]
	c.pc++

	if c.cycles == 1 {
		c.cur = nil
	}
	c.cur = append(c.cur, r)

	return r
}

func (c *cpu) readByte(ms, ls uint8) uint8 {
	<-c.clock
	c.cycles++
	addr := uint16(ms)<<8 + uint16(ls)
	return c.ram[addr]
}

func (c *cpu) writeByte(ms, ls, b uint8) {
	<-c.clock
	c.cycles++
	addr := uint16(ms)<<8 + uint16(ls)
	c.ram[addr] = b
}

// hl = ab + xy
func (c *cpu) addWord(a, b, x, y uint8) (uint8, uint8) {
	l := c.add(b, y)
	h := c.addC(a, x)
	return h, l
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

func (c *cpu) disableInterrupts() {
	//TODO: implement
}

func (c *cpu) enableInterrupts() {
	//TODO: implement
}

func (c *cpu) rr(n uint8) uint8 {
	// TODO: verify
	r := n >> 1
	c.f = 0
	if r == 0 {
		c.f |= flagZSet
	}
	return r
}

func (c *cpu) rlc(n uint8) uint8 {
	// TODO: verify
	<-c.clock
	c.cycles++
	r := n<<1 + n>>7
	c.f = 0
	if r == 0 {
		c.f |= flagZSet
	}
	if n>>7 == 1 {
		c.f |= flagCSet
	}
	return r
}

func (c *cpu) xor(a, b uint8) uint8 {
	r := a ^ b
	c.f = 0
	if r == 0 {
		c.f |= flagZSet
	}
	return r
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

func (c *cpu) pushNoClock(b uint8) {
	c.ram[c.sp] = b
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
*/

/*
// load next instruction into c.inst
// c.pc is updated
func (c *cpu) fetchInstruction() {
	opcode := c.mc.readByte(c.pc)
	c.inst = newInstruction(opcode)
	for i := uint8(0); i < c.commands[opcode].b; i++ {
		c.inst = append(c.inst, c.mc.readByte(c.pc))
	}
}

func (c *cpu) executeInstruction() {
	c.commands.execute(c)
}
*/

func (c *cpu) fetch() {
	opcode := c.mc.readByte(c.pc)
	c.pc++
	command := commandTable[opcode]
	inst := newInstruction(opcode)

	for i := uint8(0); i < command.b; i++ {
		inst = append(inst, c.mc.readByte(c.pc))
		c.pc++
	}
	c.inst = inst
	fmt.Println(c.inst)
}

func (c *cpu) execute() {
	opcode := c.inst[0]
	c.commands[opcode].f(c)
}

func (c *cpu) loop() {
	c.fetch()   // load next instruction into c.inst
	c.execute() // execute c.inst instruction
	/*
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
		case 0x07: // RLC A
			c.a = c.rlc(c.a)
		case 0x08: // LD (nn), SP
			l := c.readPc()
			h := c.readPc()
			c.writeByte(h, l, uint8(c.sp&0xFF))
			h, l = increment(h, l)
			c.writeByte(h, l, uint8(c.sp>>8))
		case 0x09: // ADD HL, BC
			c.h, c.l = c.addWord(c.h, c.l, c.b, c.c)
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
		case 0x18: // JR n
			n := int8(c.readPc())
			if n < 0 {
				c.pc -= uint16(-n)
			} else {
				c.pc += uint16(n)
			}
		case 0x19: // ADD HL, DE
			c.h, c.l = c.addWord(c.h, c.l, c.d, c.e)
		case 0x1A: // LD A, (DE)
			c.a = c.readByte(c.d, c.e)
		case 0x1E: // LD E, n
			c.e = c.readPc()
		case 0x1F: // RRA
			c.a = c.rr(c.a)
		case 0x20: // JR NZ, *
			n := c.readPc()
			if c.f&flagZSet == 0 {
				c.pc += uint16(n)
			}
		case 0x21: // LD HL, nn
			c.l = c.readPc()
			c.h = c.readPc()
		case 0x22: // LDI (HL), A
			c.writeByte(c.h, c.l, c.a)
			c.h, c.l = increment(c.h, c.l)
		case 0x26: // LD H, n
			c.h = c.readPc()
		case 0x28: // JR Z, *
			n := c.readPc()
			if c.f&flagZSet == flagZSet {
				c.pc += uint16(n)
			}
		case 0x29: // ADD HL, HL
			c.h, c.l = c.addWord(c.h, c.l, c.h, c.l)
		case 0x2A: // LDI A, (HL)
			c.a = c.readByte(c.h, c.l)
			c.h, c.l = increment(c.h, c.l)
		case 0x2E: // LD L, n
			c.l = c.readPc()
		case 0x30: // JR NC, *
			n := c.readPc()
			if c.f&flagCSet == 0 {
				c.pc += uint16(n)
			}
		case 0x31: // LD SP, nn
			l := c.readPc()
			h := c.readPc()
			c.sp = uint16(h)<<8 + uint16(l)
		case 0x32: // LDD (HL), A
			c.writeByte(c.h, c.l, c.a)
			c.h, c.l = decrement(c.h, c.l)
		case 0x36: // LD (HL), n
			c.writeByte(c.h, c.l, c.readPc())
		case 0x38: // JR C, *
			n := c.readPc()
			if c.f&flagCSet == flagCSet {
				c.pc += uint16(n)
			}
		case 0x39: // ADD HL, SP
			c.h, c.l = c.addWord(c.h, c.l, uint8(c.sp>>8), uint8(c.sp&0xFF))
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
		case 0xA8: // XOR B
			c.a = c.xor(c.a, c.b)
		case 0xA9: // XOR C
			c.a = c.xor(c.a, c.c)
		case 0xAA: // XOR D
			c.a = c.xor(c.a, c.d)
		case 0xAB: // XOR E
			c.a = c.xor(c.a, c.e)
		case 0xAC: // XOR H
			c.a = c.xor(c.a, c.h)
		case 0xAD: // XOR L
			c.a = c.xor(c.a, c.l)
		case 0xAE: // XOR (HL)
			c.a = c.xor(c.a, c.readByte(c.h, c.l))
		case 0xAF: // XOR A
			c.a = c.xor(c.a, c.a)
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
		case 0xB8: // CP B
			c.sub(c.a, c.b)
		case 0xB9: // CP B
			c.sub(c.a, c.b)
		case 0xBA: // CP C
			c.sub(c.a, c.c)
		case 0xBB: // CP D
			c.sub(c.a, c.d)
		case 0xBC: // CP E
			c.sub(c.a, c.e)
		case 0xBD: // CP H
			c.sub(c.a, c.h)
		case 0xBE: // CP L
			c.sub(c.a, c.l)
		case 0xBF: // CP (HL)
			c.sub(c.a, c.readByte(c.h, c.l))
		case 0xC1: // POP BC
			c.c = c.pop()
			c.b = c.pop()
		case 0xC3: // JP nn
			l := c.readPc()
			h := c.readPc()
			c.pc = uint16(h)<<8 + uint16(l)
		case 0xC5: // PUSH BC
			c.push(c.b)
			c.push(c.c)
			<-c.clock
			c.cycles++
		case 0xC6: // ADD A, #
			c.a = c.add(c.a, c.readPc())
		case 0xCD: // CALL nn
			l := c.readPc()
			h := c.readPc()
			c.pushNoClock(uint8(c.pc >> 8))
			c.pushNoClock(uint8(c.pc & 0xFF))
			c.pc = uint16(h)<<8 + uint16(l)
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
		case 0xEE: // XOR #
			n := c.readPc()
			c.a = c.xor(c.a, n)
		case 0xF0: // LDH A, (n)
			n := c.readPc()
			c.a = c.readByte(0xFF, n)
		case 0xF1: // POP AF
			c.f = c.pop()
			c.a = c.pop()
		case 0xF2: // LD A, (C)
			c.a = c.readByte(0xFF, c.c)
		case 0xF3: // DI
			c.di = 3 // disable interrupts after next instruction
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
		case 0xFB: // EI
			c.ei = 3 // enable interrupts after next instruction
		case 0xFE: // CP #
			c.sub(c.a, c.readPc())
		default:
			panic(fmt.Sprintf("unknown opcode 0x%02X", opcode))
		}

		if c.di > 0 {
			c.di--
		}
		if c.di == 1 {
			c.disableInterrupts()
		}

		if c.ei > 0 {
			c.ei--
		}
		if c.ei == 1 {
			c.enableInterrupts()
		}

	*/
	// reset clocks
	c.m = 0
	c.t = 0
}
