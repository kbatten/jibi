package jibi

import (
	"fmt"
	"time"
)

// A Cpu is the central proecessing unit. This one is similar to a z80. Its
// purpose is to handle interrupts, fetch and execute instructions, and
// manage the clock.
type Cpu struct {
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
	clock *Clock
	m     ClockType // machine cycles
	t     ClockType // clock cycles
	div   Word

	// current instruction buffer
	inst instruction

	// interrupt master enable
	ime            Bit
	imeDisableNext uint8

	// timers
	tima timer

	// memory
	mmu Mmu

	// internal state
	bios         []Byte
	biosFinished bool

	// notifications
	notifyInst []chan string

	// cpu information
	hz     float64
	period time.Duration
}

// NewCpu creates a new Cpu with mmu connection.
func NewCpu(mmu Mmu, bios []Byte) *Cpu {
	// use internal clock
	// 1 machine cycle = 4 clock cycles
	// machine cycles: 1.05MHz nop: 1 cycle
	// clock cycles: 4.19MHz nop: 4 cycles
	hz := 4.194304 * 1e6 / 4.0 // 4.19MHz clock to 1.05 machine cycles
	period := time.Duration(1e9 / hz)

	f := newFlagsRegister8()
	a := newRegister8(&f)
	c := newRegister8(nil)
	b := newRegister8(&c)
	e := newRegister8(nil)
	d := newRegister8(&e)
	l := newRegister8(nil)
	h := newRegister8(&l)

	cpu := &Cpu{
		a: a, b: b, c: c, d: d, e: e, f: f, l: l, h: h,
		clock: NewClock(),
		ime:   Bit(1),
		mmu:   mmu,
		bios:  bios,
		hz:    hz, period: period,
	}
	return cpu
}

// return a channel that replicates internal clock
func (c *Cpu) AttachClock() chan ClockType {
	return c.clock.Attach()
}

// return a channel that will get every instruction
func (c *Cpu) AttachInstructions() chan string {
	// TODO: use a construct like AttachClock
	inst := make(chan string)
	c.notifyInst = append(c.notifyInst, inst)
	return inst

}

func (c *Cpu) String() string {
	return fmt.Sprintf(`%s
a:%s f:%s b:%s c:%s d:%s e:%s h:%s l:%s sp:%d pc:%d
ime:%d div:0x%04X %s`,
		c.inst, c.a, c.f, c.b, c.c, c.d, c.e, c.h, c.l, c.sp, c.pc,
		c.ime, c.div, c.f.flagsString())
}

func (c *Cpu) readByte(addr Word) Byte {
	if !c.biosFinished && addr <= 0xFF {
		return c.bios[addr]
	}
	return c.mmu.ReadByteAt(addr)
}

func (c *Cpu) writeByte(addr Word, b Byte) {
	c.mmu.WriteByteAt(addr, b)
}

func (c *Cpu) readWord(addr Word) Word {
	if !c.biosFinished && addr <= 0xFF {
		return BytesToWord(c.bios[addr+1], c.bios[addr])
	}
	return c.mmu.ReadWordAt(addr)
}

func (c *Cpu) writeWord(addr Word, w Word) {
	c.mmu.WriteWordAt(addr, w)
}

func (c *Cpu) fetch() {
	op := opcode(c.readByte(c.pc))
	c.pc++
	if op == 0xCB {
		op = opcode(0xCB00 + uint16(c.readByte(c.pc)))
		c.pc++
	}
	if cmd, ok := commandTable[op]; ok {
		p := []Byte{}
		for i := uint8(0); i < cmd.b; i++ {
			p = append(p, c.readByte(c.pc))
			c.pc++
		}
		c.inst = newInstruction(op, p...)
	} else {
		panic(op)
	}
}

func (c *Cpu) execute() {
	cmd := commandTable[c.inst.o]
	cmd.f(c)
	c.t += cmd.t
	c.m += cmd.t * 4

	// handle disabling interrupts one instruction after DI
	if c.imeDisableNext > 0 {
		c.imeDisableNext--
		if c.imeDisableNext == 0 {
			c.ime = Bit(0)
		}
	}
}

// setInterrupt sets the specific interrupt.
func (cpu *Cpu) setInterrupt(in Interrupt) {
	cpu.mmu.SetInterrupt(in)
}

// resetInterrupt resets the specific interrupt.
func (cpu *Cpu) resetInterrupt(in Interrupt) {
	cpu.mmu.ResetInterrupt(in)
}

// getInterrupt returns the highest priority enabled interrupt.
func (cpu *Cpu) getInterrupt() Interrupt {
	ie := cpu.readByte(AddrIE)
	interrupts := cpu.readByte(AddrIF)

	if Byte(InterruptVblank)&ie&interrupts != 0 {
		return InterruptVblank
	} else if Byte(InterruptLCDC)&ie&interrupts != 0 {
		return InterruptLCDC
	} else if Byte(InterruptTimer)&ie&interrupts != 0 {
		return InterruptTimer
	} else if Byte(InterruptSerial)&ie&interrupts != 0 {
		return InterruptSerial
	} else if Byte(InterruptKeypad)&ie&interrupts != 0 {
		return InterruptKeypad
	}
	return 0
}

func (cpu *Cpu) io() {
	iflags := cpu.mmu.ReadIoByte(AddrIF)
	if cpu.ime == 0 {
		iflags = 0 // mask all interrupts
	} else {
		ie := cpu.readByte(AddrIE)
		iflags &= ie // mask interrupts
	}
	cpu.writeByte(AddrIF, iflags)
}

func (cpu *Cpu) interrupt() {
	if cpu.ime == 1 {
		in := cpu.getInterrupt()
		if in > 0 {
			cpu.ime = 0
			cpu.push(cpu.pc)
			cpu.jp(in.Address())
			cpu.resetInterrupt(in)
		}
	}
}

type timer struct {
	v       Byte
	div     uint16
	running bool
}

func newTimer() *timer {
	return &timer{}
}

func (t *timer) run(c ClockType, f Byte, tma Byte) (Byte, bool) {
	overflow := false

	tmaBit := uint16(1)
	if tma == 0x00 {
		tmaBit = 0x0400 // 10th bit
	} else if tma == 0x01 {
		tmaBit = 0x0010 // 4th bit
	} else if tma == 0x02 {
		tmaBit = 0x0040 // 6th bit
	} else if tma == 0x03 {
		tmaBit = 0x0100 // 8th bit
	}

	p := t.div & tmaBit
	t.div += uint16(c)
	if p == 0 { // previously 0
		if t.div&tmaBit == tmaBit { // now 1
			t.v += 1
			if t.v == 0 {
				overflow = true
			}
		}
	}

	return t.v, overflow
}

func (t *timer) stop() {
	t.v = 0
	t.div = 0
	t.running = false
}

func (cpu *Cpu) timers() {
	// update divider
	div := cpu.readByte(AddrDIV)
	cpu.div = (cpu.div & 0x00FF) | (Word(div) << 8)
	cpu.div += Word(cpu.t)
	div = Byte(cpu.div >> 8)
	cpu.mmu.WriteElevatedByteAt(AddrDIV, div)

	// update timer
	tac := cpu.readByte(AddrTAC)
	if tac&0x04 == 0x00 {
		cpu.tima.stop()
		return
	}
	tima := cpu.readByte(AddrTIMA)
	tma := cpu.readByte(AddrTMA)

	tima, interrupt := cpu.tima.run(cpu.t, tac&0x03, tma)
	if interrupt {
		cpu.setInterrupt(InterruptTimer)
	}
	cpu.writeByte(AddrTIMA, tima)
}

func (c *Cpu) step(first bool, t uint32) (CommanderStateFn, bool, uint32, uint32) {
	// reset clocks
	c.m = 0
	c.t = 0
	if !c.biosFinished && c.pc == 0x0100 {
		c.biosFinished = true
	}

	c.io()        // handle memory mapped io
	c.interrupt() // handle interrupts
	c.fetch()     // load next instruction into c.inst
	c.execute()   // execute c.inst instruction
	c.timers()    // handle tima, tma, tac

	c.clock.AddCycles(c.t)

	for _, inst := range c.notifyInst {
		inst <- c.inst.String()
	}

	return c.step, false, 0, 0
}
