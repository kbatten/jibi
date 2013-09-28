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
	inst instruction

	mc mmu // read/write bytes and words
}

const (
	flagZ = 0x80
	flagN = 0x40
	flagH = 0x20
	flagC = 0x10
)

func newCpu(cart cartridge) *cpu {
	// use internal clock
	// 1 machine cycle = 4 clock cycles
	// machine cycles: 1.05MHz nop: 1 cycle
	// clock cycles: 4.19MHz nop: 4 cycles
	hz := 4.194304 * 1e6 / 4.0 // 4.19MHz clock to 1.05 machine cycles
	period := time.Duration(1e9 / hz)
	ticker := time.NewTicker(period)
	clock := ticker.C

	// internal mmu
	mc := newMmu(cart)

	f := newFlagsRegister8()
	a := newRegister8(&f)
	c := newRegister8(nil)
	b := newRegister8(&c)
	e := newRegister8(nil)
	d := newRegister8(&e)
	l := newRegister8(nil)
	h := newRegister8(&l)

	return &cpu{a: a, b: b, c: c, d: d, e: e, f: f, l: l, h: h,
		sp: 0xFFFE, mTicker: ticker, mClock: clock, mc: mc}
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
}

func (c *cpu) xor(a, b uint8) uint8 {
	r := a ^ b
	c.f.set(0)
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	return r
}

func (c *cpu) and(a, b uint8) uint8 {
	r := a & b
	c.f.set(0)
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	c.f.setFlag(flagH)
	return r
}

func (c *cpu) or(a, b uint8) uint8 {
	r := a | b
	c.f.set(0)
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	return r
}

func (c *cpu) inc(a uint8) uint8 {
	r := a + 1
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	c.f.setFlag(flagN)
	if a&0x0F == 0x0F {
		c.f.setFlag(flagH)
	}
	return r
}

func (c *cpu) dec(a uint8) uint8 {
	r := a - 1
	if r == 0 {
		c.f.resetFlag(flagZ)
	}
	c.f.setFlag(flagN)
	if a&0x0F == 0x0F {
		c.f.setFlag(flagH)
	}
	return r
}

func (c *cpu) subC(a, b uint8) uint8 {
	carry := uint8(0)
	if c.f.getFlag(flagC) {
		carry = 1
	}
	return c.sub(a, b+carry)
}

func (c *cpu) sub(a, b uint8) uint8 {
	r := a - b
	c.f.set(0)
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	c.f.setFlag(flagN)
	if a&0x0F < b&0x0F {
		c.f.setFlag(flagH)
	}
	if a < b {
		c.f.setFlag(flagC)
	}
	return r
}

func (c *cpu) addWordR(a uint16, b int8) uint16 {
	h := uint8(a >> 8)
	l := uint8(a & 0xFF)
	bu := uint8(b)
	if b < 0 {
		bu = uint8(-b)
		l = c.sub(l, bu)
		h = c.subC(h, 0)
		return uint16(h)<<8 + uint16(l)
	}
	l = c.add(l, bu)
	h = c.addC(h, 0)
	return uint16(h)<<8 + uint16(l)
}

func (c *cpu) addC(a, b uint8) uint8 {
	carry := uint8(0)
	if c.f.getFlag(flagC) {
		carry = 1
	}
	return c.add(a, b+carry)
}

func (c *cpu) add(a, b uint8) uint8 {
	r := a + b
	c.f.set(0)
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	if a&0x0F+b&0x0F > 0x0F {
		c.f.setFlag(flagH)
	}
	if uint16(a)+uint16(b) > 0xFF {
		c.f.setFlag(flagC)
	}
	return r
}

func (c *cpu) rr(n uint8) uint8 {
	r := n >> 1
	if c.f.getFlag(flagC) { // old carry is bit 7
		r += 1 << 7
	}
	c.f.set(0)
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	if n&0x01 == 0x01 { // carry is old bit 0
		c.f.setFlag(flagC)
	}
	return r
}

func (c *cpu) jrF(f uint8, n int8) {
	if c.f.getFlag(f) == true {
		c.jr(n)
	}
}

func (c *cpu) jrNF(f uint8, n int8) {
	if c.f.getFlag(f) == false {
		c.jr(n)
	}
}

func (c *cpu) jr(n int8) {
	if n < 0 {
		c.pc += register16(-n)
		return
	}
	c.pc += register16(n)
}

func (c *cpu) jp(addr address) {
	c.pc = register16(addr)
}

func (c *cpu) call(addr address) {
	c.pushWord(uint16(c.pc))
	c.jp(addr)
}

func (c *cpu) popWord() uint16 {
	r := c.mc.readWord(address(c.sp))
	c.sp += 2
	return r
}

func (c *cpu) pushWord(w uint16) {
	c.mc.writeWord(address(c.sp-2), w)
	c.sp -= 2
}

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
}

func (c *cpu) execute() {
	opcode := c.inst[0]
	commandTable[opcode].f(c)
}

func (c *cpu) loop() {
	c.fetch()   // load next instruction into c.inst
	c.execute() // execute c.inst instruction

	// reset clocks
	c.m = 0
	c.t = 0
}
