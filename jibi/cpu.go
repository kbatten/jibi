package jibi

import (
	"fmt"
	"time"
)

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
	tClock chan uint32 // clock cycles since last
	m      uint8       // machine cycles
	t      uint8       // clock cycles

	// current instruction buffer
	inst instruction

	// interrupt master enable
	ime Bit

	mm MmuConnection // read/write bytes and words

	// cpu information
	hz     float64
	period time.Duration
}

func NewCpu() *Cpu {
	// use internal clock
	// 1 machine cycle = 4 clock cycles
	// machine cycles: 1.05MHz nop: 1 cycle
	// clock cycles: 4.19MHz nop: 4 cycles
	hz := 4.194304 * 1e6 / 4.0 // 4.19MHz clock to 1.05 machine cycles
	period := time.Duration(1e9 / hz)
	clock := make(chan uint32)

	f := newFlagsRegister8()
	a := newRegister8(&f)
	c := newRegister8(nil)
	b := newRegister8(&c)
	e := newRegister8(nil)
	d := newRegister8(&e)
	l := newRegister8(nil)
	h := newRegister8(&l)

	return &Cpu{a: a, b: b, c: c, d: d, e: e, f: f, l: l, h: h,
		tClock: clock, ime: Bit(1),
		hz: hz, period: period,
	}
}

func (c *Cpu) String() string {
	return fmt.Sprintf(`%v
    a:%v b:%v c:%v d:%v e:%v f:%v h:%v l:%v
    af:0x%04X bc:0x%04X de:0x%04X hl:0x%04X sp:%v pc:%v
	%s`,
		c.inst, c.a, c.b, c.c, c.d, c.e, c.f, c.h, c.l,
		c.a.Uint16(), c.b.Uint16(), c.d.Uint16(), c.h.Uint16(), c.sp, c.pc,
		c.f.flagsString())
}

func (c *Cpu) ConnectMmu(m *Mmu) {
	c.mm = m.Connect()
}

func (c *Cpu) reset() {
	c.a.reset()
	c.b.reset()
	c.c.reset()
	c.d.reset()
	c.e.reset()
	c.f.reset()
	c.h.reset()
	c.l.reset()
	c.sp = 0x0000
	c.pc = 0x0000
	c.m = 0
	c.t = 0
	c.mm.writeByte(Word(0xFFFF), Byte(0xFF))
	c.ime = 1
}

// z reset
// n reset
// h and c set or reset according to operation
func (c *Cpu) addWordR(a Worder, b Byter) Word {
	h := a.High()
	l := a.Low()
	bi := int8(b.Uint8())
	if bi < 0 {
		b = Byte(uint8(-bi))
		l = c.sub(l, b)
		h = c.sbc(h, Byte(0))
		c.f.resetFlag(flagZ)
		c.f.resetFlag(flagN)
		return bytesToWord(h, l)
	}
	l = c.add(l, b)
	h = c.adc(h, Byte(0))
	c.f.resetFlag(flagZ)
	c.f.resetFlag(flagN)
	return bytesToWord(h, l)
}

func (c *Cpu) fetch() {
	op := opcode(c.mm.readByte(c.pc))
	c.pc++
	if op == 0xCB {
		op = opcode(0xCB00 + uint16(c.mm.readByte(c.pc)))
		c.pc++
	}
	command := commandTable[op]
	c.inst = newInstruction(op)

	for i := uint8(0); i < command.b; i++ {
		c.inst.p = append(c.inst.p, c.mm.readByte(c.pc))
		c.pc++
	}
}

func (c *Cpu) execute() {
	if c.pc == 0x0100 {
		c.mm.unloadBios()
	}
	if cmd, ok := commandTable[c.inst.o]; ok {
		cmd.f(c)
		c.t += cmd.t
		c.m += cmd.t * 4
	}
}

func (c *Cpu) getInterrupt() interrupt {
	iereg := c.mm.readByte(Word(0xFFFF)) // interrupt enable
	iflag := c.mm.readByte(Word(0xFF0F)) // interrupt flags
	if Byte(interruptVBlank)&iereg&iflag != 0 {
		return interruptVBlank
	} else if Byte(interruptLCDC)&iereg&iflag != 0 {
		return interruptLCDC
	} else if Byte(interruptTimer)&iereg&iflag != 0 {
		return interruptTimer
	} else if Byte(interruptSerial)&iereg&iflag != 0 {
		return interruptSerial
	} else if Byte(interruptKeypad)&iereg&iflag != 0 {
		return interruptKeypad
	}
	return 0
}

func (c *Cpu) resetInterrupt(i interrupt) {
	iflag := c.mm.readByte(Word(0xFF0F))
	iflag &= (Byte(i) ^ 0xFF)
	c.mm.writeByte(Word(0xFF0F), iflag)
}

// memoryDevice and flag handler
type interruptFlags struct {
	v *uint8
}

func newInterruptFlags() interruptFlags {
	return interruptFlags{new(uint8)}
}

func (i interruptFlags) readByte(addr Worder) uint8 {
	return *i.v
}

func (i interruptFlags) writeByte(addr Worder, b uint8) {
	*i.v = b
}

func (i interruptFlags) set(in interrupt) {
	*i.v |= uint8(in)
}

type interrupt uint8

const (
	interruptVBlank interrupt = 0x01 << iota
	interruptLCDC
	interruptTimer
	interruptSerial
	interruptKeypad
)

func (i interrupt) Word() Word {
	switch i {
	case interruptVBlank:
		return Word(0x0040)
	case interruptLCDC:
		return Word(0x0048)
	case interruptTimer:
		return Word(0x0050)
	case interruptSerial:
		return Word(0x0058)
	case interruptKeypad:
		return Word(0x0060)
	default:
		return Word(0)
	}
}

func (i interrupt) String() string {
	switch i {
	case interruptVBlank:
		return "VBlank"
	case interruptLCDC:
		return "LCDC"
	case interruptTimer:
		return "Timer"
	case interruptSerial:
		return "Serial"
	case interruptKeypad:
		return "Keypad"
	default:
		return "UNKNOWN"
	}
}

func (c *Cpu) interrupt() {
	if c.ime == 1 {
		c.ime = 0
		in := c.getInterrupt()
		if in > 0 {
			panic("interrupt")
			c.push(c.pc)
			c.jp(in.Word())
			c.resetInterrupt(in)
		}
	}
}

func (c *Cpu) Step() uint8 {
	// reset clocks
	c.m = 0
	c.t = 0
	c.interrupt() // handle interrupts
	c.fetch()     // load next instruction into c.inst
	c.execute()   // execute c.inst instruction

	return c.t
}
