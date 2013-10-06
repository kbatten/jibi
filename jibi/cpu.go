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

	mmu MemoryCommander

	zero   MemoryDevice
	iflags Byte
	ie     Byte

	// internal state
	biosFinished bool

	// cpu information
	hz     float64
	period time.Duration
}

// NewCpu creates a new Cpu with mmu connection.
func NewCpu(mmu MemoryCommander) *Cpu {
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

	commander := NewCommander("cpu")

	cpu := &Cpu{CommanderInterface: commander,
		a: a, b: b, c: c, d: d, e: e, f: f, l: l, h: h,
		ime:    Bit(1),
		mmu:    mmu,
		zero:   NewRamDevice(Word(0xFF80), Word(0x7F), nil),
		iflags: Byte(0),
		ie:     Byte(0),
		hz:     hz, period: period,
	}
	cmdHandlers := map[Command]CommandFn{
		CmdClockAccumulator: cpu.cmdClock,
		CmdString:           cpu.cmdString,
		CmdSetInterrupt:     cpu.cmdSetInterrupt,
		CmdOnInstruction:    cpu.cmdOnInstruction,
	}

	commander.start(cpu.step, cmdHandlers, nil)
	mmu.RunCommand(CmdHandleMemory, MemoryHandlerRequest{0xFF80, 0xFFFE, cpu})
	mmu.RunCommand(CmdHandleMemory, MemoryHandlerRequest{AddrIE, AddrIE, cpu})
	mmu.RunCommand(CmdHandleMemory, MemoryHandlerRequest{AddrIF, AddrIF, cpu})
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
	if resp, ok := resp.(chan chan ClockType); !ok {
		panic("invalid command response type")
	} else {
		clk := make(chan ClockType)
		c.tClocks = append(c.tClocks, NewClock(clk))
		resp <- clk
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
a:%s b:%s c:%s d:%s e:%s f:%s h:%s l:%s
af:0x%04X bc:0x%04X de:0x%04X hl:0x%04X sp:%s pc:%s
%s`,
		c.inst, c.a, c.b, c.c, c.d, c.e, c.f, c.h, c.l,
		c.a.Word(), c.b.Word(), c.d.Word(), c.h.Word(), c.sp, c.pc,
		c.f.flagsString())
}

func (c *Cpu) String() string {
	resp := make(chan string)
	c.RunCommand(CmdString, resp)
	return <-resp
}

// ReadByteAt reads a single byte from the cpu at the specified address.
func (c *Cpu) ReadByteAt(addr Worder) Byte {
	req := ReadByteAtReq{addr.Word(), make(chan Byte)}
	c.RunCommand(CmdReadByteAt, req)
	return <-req.b
}

// WriteByteAt writes a single byte to the cpu at the specified address.
func (c *Cpu) WriteByteAt(addr Worder, b Byter) {
	req := WriteByteAtReq{addr.Word(), b.Byte()}
	c.RunCommand(CmdWriteByteAt, req)
}

func (c *Cpu) readByte(addr Worder) Byte {
	a := addr.Word()
	if 0xFF80 <= a && a <= 0xFFFE {
		return c.zero.ReadByteAt(addr)
	} else if AddrIF == a {
		panic("IF")
		return c.iflags
	} else if AddrIE == a {
		panic("IE")
		return c.ie
	}
	c.yield()
	return c.mmu.ReadByteAt(a)
}

func (c *Cpu) writeByte(addr Worder, b Byter) {
	a := addr.Word()
	if 0xFF80 <= a && a <= 0xFFFE {
		c.zero.WriteByteAt(addr, b)
	} else if AddrIF == a {
		panic("IF")
		c.iflags = b.Byte()
	} else if AddrIE == a {
		panic("IE")
		c.ie = b.Byte()
	} else {
		c.yield()
		c.mmu.WriteByteAt(addr, b)
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
	if !c.biosFinished && c.pc == 0x0100 {
		c.mmu.RunCommand(CmdUnloadBios, nil)
		c.biosFinished = true
	}
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
	c.interrupt() // handle interrupts
	c.fetch()     // load next instruction into c.inst
	c.execute()   // execute c.inst instruction

	for _, clk := range c.tClocks {
		clk.AddCycles(c.t)
	}
	return c.step, false, 0, 0
}
