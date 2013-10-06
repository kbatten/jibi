package jibi

import (
	"fmt"
	"time"
)

// A Cpu is the central proecessing unit. This one is similar to a z80. Its
// purpose is to handle interrupts, fetch and execute instructions, and
// manage the clock.
type Cpu struct {
	Commander

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
	tClocks []*clock // t clock cycle exported clocks
	m       uint8    // machine cycles
	t       uint8    // clock cycles

	// current instruction buffer
	inst instruction

	// interrupt master enable
	ime Bit

	mmu MemoryCommander
	irq *Irq

	zero MemoryDevice

	// internal state
	biosFinished bool

	// cpu information
	hz     float64
	period time.Duration
}

// NewCpu creates a new Cpu with mmu and irq connections.
func NewCpu(mmu MemoryCommander, irq *Irq) *Cpu {
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

	cpu := &Cpu{Commander: commander,
		a: a, b: b, c: c, d: d, e: e, f: f, l: l, h: h,
		ime: Bit(1),
		mmu: mmu, irq: irq,
		zero: NewRamDevice(Word(0xFF80), Word(0x7F), nil),
		hz:   hz, period: period,
	}
	cmdHandlers := map[Command]CommandFn{
		CmdClock:  cpu.cmdClock,
		CmdString: cpu.cmdString,
	}
	commander.Start(cpu.step, cmdHandlers, nil)

	handler := MemoryHandlerRequest{0xFF80, 0xFFFE, cpu}
	mmu.RunCommand(CmdHandleMemory, handler)
	return cpu
}

func (c *Cpu) cmdClock(resp interface{}) {
	if resp, ok := resp.(chan chan ClockType); !ok {
		panic("invalid command response type")
	} else {
		clk := make(chan ClockType, 1)
		c.tClocks = append(c.tClocks, newClock(clk))
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
		c.a.Byte(), c.b.Byte(), c.d.Byte(), c.h.Byte(), c.sp, c.pc,
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
	}
	return c.mmu.ReadByteAt(a)
}

func (c *Cpu) writeByte(addr Worder, b Byter) {
	a := addr.Word()
	if 0xFF80 <= a && a <= 0xFFFE {
		c.zero.WriteByteAt(addr, b)
	} else {
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

// A ClockType is simply the type used for all clocks
type ClockType uint32

// a clock sends number of clock cycle since last successful send
// so if a non-blocking send fails, the cycles accumulate
// on successful send the cycles is reset
// sends happen on machine cycle end
type clock struct {
	v ClockType
	c chan ClockType
}

func newClock(c chan ClockType) *clock {
	return &clock{ClockType(0), c}
}

func (c *clock) addCycles(cycles uint8) {
	c.v += ClockType(cycles)
	//v := uint8(c.v)
	//if c.v > 255 {
	//	v = 255
	//}

	select {
	case c.c <- c.v:
		//c.v -= ClockType(v)
		c.v = 0
	default:
	}
}

// Clock returns a new channel that holds acumulating clock ticks.
func (c *Cpu) Clock() chan ClockType {
	resp := make(chan chan ClockType)
	c.RunCommand(CmdClock, resp)
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
	c.inst = newInstruction(op)

	for i := uint8(0); i < command.b; i++ {
		c.inst.p = append(c.inst.p, c.readByte(c.pc))
		c.pc++
	}
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
		c.ime = 0
		in := c.irq.GetInterrupt()
		if in > 0 {
			c.push(c.pc)
			c.jp(in.Address())
			c.irq.ResetInterrupt(in)
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
		clk.addCycles(c.t)
	}

	return c.step, false, 0, 0
}
