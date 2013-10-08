package jibi

import (
	"fmt"
	"time"
)

// A Cpu is the central proecessing unit. This one is similar to a z80. Its
// purpose is to handle interrupts, fetch and execute instructions, and
// manage the clock.
type Cpu struct {
	CommanderInterface

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
	tClocks []*Clock // t clock cycle exported clocks
	m       uint8    // machine cycles
	t       uint8    // clock cycles

	// current instruction buffer
	inst instruction

	// interrupt master enable
	ime Bit

	mmu *Mmu

	bios   []Byte
	zero   []Byte
	iflags Byte
	ie     Byte
	tma    Byte

	// internal state
	biosFinished bool

	// notifications
	notifyInst []chan string

	// cpu information
	hz     float64
	period time.Duration
}

// NewCpu creates a new Cpu with mmu connection.
func NewCpu(mmu *Mmu, bios []Byte) *Cpu {
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

	biosFinished := true
	if len(bios) > 0 {
		biosFinished = false
		biosN := make([]Byte, 0x100)
		copy(biosN, bios)
		bios = biosN
	}

	commander := NewCommander("cpu")
	cpu := &Cpu{CommanderInterface: commander,
		a: a, b: b, c: c, d: d, e: e, f: f, l: l, h: h,
		ime:  Bit(1),
		mmu:  mmu,
		bios: bios,
		zero: make([]Byte, 0x7F),
		hz:   hz, period: period,
		biosFinished: biosFinished,
	}
	cmdHandlers := map[Command]CommandFn{
		CmdClockAccumulator: cpu.cmdClock,
		CmdString:           cpu.cmdString,
		CmdSetInterrupt:     cpu.cmdSetInterrupt,
		CmdOnInstruction:    cpu.cmdOnInstruction,
		CmdHandleMemory: func(r interface{}) {
			if mmu != nil {
				mmu.cmdHandleMemory(r)
			}
		},
		CmdHandleCpuMemory: func(r interface{}) {
			if mmu != nil {
				mmu.cmdHandleCpuMemory(r)
			}
		},
	}

	commander.start(cpu.step, cmdHandlers, nil)
	if mmu != nil {
		mmu.connectCpu(cpu)
		mmu.handleLocalMemory(CpuMemoryHandler{0xFF80, 0xFFFE, cpu})
		mmu.handleLocalMemory(CpuMemoryHandler{AddrIE, AddrIE, cpu})
		mmu.handleLocalMemory(CpuMemoryHandler{AddrIF, AddrIF, cpu})
		mmu.handleLocalMemory(CpuMemoryHandler{AddrTMA, AddrTMA, cpu})
	}
	return cpu
}

func (c *Cpu) cmdClock(resp interface{}) {
	if resp, ok := resp.(chan chan ClockType); !ok {
		panic("invalid command response type")
	} else {
		clk := make(chan ClockType, 1)
		c.tClocks = append(c.tClocks, NewClock(clk))
		resp <- clk
	}
}

func (c *Cpu) cmdOnInstruction(resp interface{}) {
	if resp, ok := resp.(chan chan string); !ok {
		panic("invalid command response type")
	} else {
		inst := make(chan string)
		c.notifyInst = append(c.notifyInst, inst)
		resp <- inst
	}
}

func (c *Cpu) cmdString(resp interface{}) {
	if resp, ok := resp.(chan string); !ok {
		panic("invalid command response type")
	} else {
		resp <- c.str()
	}
}

func (c *Cpu) str() string {
	return fmt.Sprintf(`%s
a:%s f:%s b:%s c:%s d:%s e:%s h:%s l:%s sp:%s pc:%s
ime:%d ie:0x%02X if:0x%02X %s`,
		c.inst, c.a, c.f, c.b, c.c, c.d, c.e, c.h, c.l, c.sp, c.pc,
		c.ime, c.ie, c.iflags, c.f.flagsString())
}

func (c *Cpu) String() string {
	resp := make(chan string)
	c.RunCommand(CmdString, resp)
	return <-resp
}

// ReadByteAt reads a single byte from the cpu at the specified address.
func (c *Cpu) ReadByteAt(addr Worder, b chan Byte) {
	req := ReadByteAtReq{addr.Word(), b}
	c.RunCommand(CmdReadByteAt, req)
}

// WriteByteAt writes a single byte to the cpu at the specified address.
func (c *Cpu) WriteByteAt(addr Worder, b Byter) {
	req := WriteByteAtReq{addr.Word(), b.Byte()}
	c.RunCommand(CmdWriteByteAt, req)
}

// ReadLocalByteAt reads a single byte from the cpu at the specified address.
func (c *Cpu) ReadLocalByteAt(addr Worder) Byte {
	a := addr.Word()
	if !c.biosFinished && a <= 0xFF {
		return c.bios[addr.Word()]
	} else if 0xFF80 <= a && a <= 0xFFFE {
		return c.zero[a-0xFF80]
	} else if AddrIF == a {
		return c.iflags
	} else if AddrIE == a {
		return c.ie
	} else if AddrTMA == a {
		return c.tma
	} else if c.mmu.isLocalMemory(addr) {
		return c.mmu.readLocalByte(addr)
	}
	panic("cpu read out of range")
}

func (c *Cpu) readByte(addr Worder) Byte {
	a := addr.Word()
	if !c.biosFinished && a <= 0xFF {
		return c.bios[a]
	} else if 0xFF80 <= a && a <= 0xFFFE {
		return c.zero[a-0xFF80]
	} else if AddrIF == a {
		panic("IF")
		return c.iflags
	} else if AddrIE == a {
		panic("IE")
		return c.ie
	} else if AddrTMA == a {
		return c.tma
	} else if c.mmu.isLocalMemory(addr) {
		return c.mmu.readLocalByte(addr)
	}
	c.yield()
	return c.mmu.readRemoteByte(a)
}

// WriteLocalByteAt writes a single byte to the cpu at the specified address.
func (c *Cpu) WriteLocalByteAt(addr Worder, b Byter) {
	a := addr.Word()
	if 0xFF80 <= a && a <= 0xFFFE {
		c.zero[a-0xFF80] = b.Byte()
	} else if AddrIF == a {
		c.iflags = b.Byte()
	} else if AddrIE == a {
		c.ie = b.Byte()
	} else if AddrTMA == a {
		c.tma = b.Byte()
	} else if c.mmu.isLocalMemory(addr) {
		c.mmu.writeLocalByte(addr, b)
	}
	panic("cpu write out of range")
}

func (c *Cpu) writeByte(addr Worder, b Byter) {
	a := addr.Word()
	if 0xFF80 <= a && a <= 0xFFFE {
		c.zero[a-0xFF80] = b.Byte()
	} else if AddrIF == a {
		c.iflags = b.Byte()
	} else if AddrIE == a {
		c.ie = b.Byte()
	} else if AddrTMA == a {
		c.tma = b.Byte()
	} else if c.mmu.isLocalMemory(addr) {
		c.mmu.writeLocalByte(addr, b)
	} else {
		c.yield()
		c.mmu.writeRemoteByte(addr, b)
	}
}

func (c *Cpu) readWord(addr Worder) Word {
	l := c.readByte(addr)
	h := c.readByte(addr.Word() + 1)
	return BytesToWord(h, l)
}

func (c *Cpu) writeWord(addr Worder, w Worder) {
	c.writeByte(addr, w.Low())
	c.writeByte(addr.Word()+1, w.High())
}

// Clock returns a new channel that holds acumulating clock ticks.
func (c *Cpu) Clock() chan ClockType {
	resp := make(chan chan ClockType)
	c.RunCommand(CmdClockAccumulator, resp)
	return <-resp
}

func (c *Cpu) fetch() {
	op := opcode(c.readByte(c.pc))
	c.pc++
	if op == 0xCB {
		op = opcode(0xCB00 + uint16(c.readByte(c.pc)))
		c.pc++
	}
	command := commandTable[op]
	p := []Byte{}
	for i := uint8(0); i < command.b; i++ {
		p = append(p, c.readByte(c.pc))
		c.pc++
	}
	c.inst = newInstruction(op, p...)
}

func (c *Cpu) execute() {
	if cmd, ok := commandTable[c.inst.o]; ok {
		cmd.f(c)
		c.t += cmd.t
		c.m += cmd.t * 4
	}
}

func (c *Cpu) interrupt() {
	if c.ime == 1 {
		in := c.getInterrupt()
		if in > 0 {
			c.ime = 0
			c.push(c.pc)
			c.jp(in.Address())
			c.resetInterrupt(in)
		}
	}
}

func (c *Cpu) step(first bool, t uint32) (CommanderStateFn, bool, uint32, uint32) {
	// reset clocks
	c.m = 0
	c.t = 0
	if !c.biosFinished && c.pc == 0x0100 {
		c.biosFinished = true
	}
	for _, inst := range c.notifyInst {
		inst <- c.str()
	}

	c.interrupt() // handle interrupts
	c.fetch()     // load next instruction into c.inst
	c.execute()   // execute c.inst instruction

	for _, clk := range c.tClocks {
		clk.AddCycles(c.t)
	}
	return c.step, false, 0, 0
}
