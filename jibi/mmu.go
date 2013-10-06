package jibi

import (
	"fmt"
)

// An Mmu is the memory management unit. Its purpose is to dispatch read and
// write requeststo the appropriate module (cpu, gpu, etc) based on the memory
// address.
type Mmu struct {
	CommanderInterface

	// memory
	bios     MemoryDevice
	handlers []memoryHandler

	// internal state
	biosFinished bool
}

// NewMmu creates a new Mmu with an optional bios that replaces 0x0000-0x00FF.
func NewMmu(bios []Byte) *Mmu {
	biosFinished := false
	if len(bios) == 0 {
		biosFinished = true
	}
	// add ram only handlers
	workingRam := NewEchoRamDevice(Word(0xC000), Word(0xE000), Word(0x2000), nil)
	handlers := []memoryHandler{
		memoryHandler{0xC000, 0xDFFF, workingRam},  // internal
		memoryHandler{0xE000, 0xFDFF, workingRam},  // echo
		memoryHandler{0xFEA0, 0xFEFF, nilModule{}}, // unusable
	}
	commander := NewCommander("mmu")
	mmu := &Mmu{
		commander,
		NewRomDevice(Word(0x0000), Word(0xFF), bios),
		handlers,
		biosFinished,
	}
	cmdHandlers := map[Command]CommandFn{
		CmdHandleMemory: mmu.cmdHandleMemory,
		CmdReadByteAt:   mmu.cmdReadByteAt,
		CmdWriteByteAt:  mmu.cmdWriteByteAt,
	}
	commander.start(nil, cmdHandlers, nil)
	return mmu
}

func (m *Mmu) cmdHandleMemory(resp interface{}) {
	if v, ok := resp.(MemoryHandlerRequest); !ok {
		panic("invalid command response type")
	} else {
		handler := memoryHandler{v.start, v.end, v.dev}
		m.handlers = append(m.handlers, handler)
	}
}

func (m *Mmu) cmdReadByteAt(resp interface{}) {
	if req, ok := resp.(ReadByteAtReq); !ok {
		panic("invalid command response type")
	} else {
		b, _ := m.readByte(req.addr)
		req.b <- b
	}
}

func (m *Mmu) cmdWriteByteAt(resp interface{}) {
	if req, ok := resp.(WriteByteAtReq); !ok {
		panic("invalid command response type")
	} else {
		m.writeByte(req.addr, req.b)
	}
}

// A MemoryDevice provides random 16bit read and write access.
type MemoryDevice interface {
	ReadByteAt(Worder) Byte
	WriteByteAt(Worder, Byter)
}

// A MemoryHandlerRequest holds the info needed to map an address to a
// MemoryDevice.
type MemoryHandlerRequest struct {
	start Word
	end   Word
	dev   MemoryDevice
}

type memoryHandler struct {
	start Word
	end   Word
	dev   MemoryDevice
}

func (m *Mmu) selectMemoryDevice(addr Worder) (MemoryDevice, Word, error) {
	a := addr.Word()
	if 0 <= a && a < 0x0FF && !m.biosFinished {
		return m.bios, a, nil
	}
	for _, mh := range m.handlers {
		if mh.start <= a && a <= mh.end {
			return mh.dev, mh.start, nil
		}
	}
	u := "unknown"
	if a == 0xFF00 {
		//u = "Register for reading joy pad info and determining system type. (R/W)"
	} else if a == 0xFF01 {
		u = "Serial transfer data (R/W)"
	} else if a == 0xFF02 {
		u = "SIO control (R/W)"
	} else if a == 0xFF10 {
		u = "Sound Mode 1 register, Sweep register (R/W)"
	} else if a == 0xFF11 {
		u = "Sound Mode 1 register, Sound length/Wave pattern duty (R/W)"
	} else if a == 0xFF12 {
		u = "Sound Mode 1 register, Envelope (R/W)"
	} else if a == 0xFF13 {
		u = "Sound Mode 1 register, Frequency lo (W)"
	} else if a == 0xFF14 {
		u = "Sound Mode 1 register, Frequency hi (R/W)"
	} else if a == 0xFF17 {
		u = "Sound Mode 2 register, envelope (R/W)"
	} else if a == 0xFF19 {
		u = "Sound Mode 2 register, frequency"
	} else if a == 0xFF1A {
		u = "Sound Mode 3 register, Sound on/off (R/W)"
	} else if a == 0xFF21 {
		u = "Sound Mode 4 register, envelope (R/W)"
	} else if a == 0xFF23 {
		u = "Sound Mode 4 register, counter/consecutive; inital (R/W)"
	} else if a == 0xFF24 {
		u = "Channel control / ON-OFF / Volume (R/W)"
	} else if a == 0xFF25 {
		u = "Selection of Sound output terminal (R/W)"
	} else if a == 0xFF26 {
		u = "Sound on/off (R/W)"
	}
	if u == "unknown" {
		panic(fmt.Errorf("unhandled memory access: 0x%04X - %s", addr, u))
	}
	return nilModule{}, Word(0), fmt.Errorf("unhandled memory access: 0x%04X - %s", addr, u)
}

// A ReadByteAtReq holds the info needed to read a byte from a MemoryCommander.
type ReadByteAtReq struct {
	addr Word
	b    chan Byte
}

// ReadByteAt reads a single byte from the mmu at the specified address.
func (m *Mmu) ReadByteAt(addr Worder) Byte {
	req := ReadByteAtReq{addr.Word(), make(chan Byte)}
	m.RunCommand(CmdReadByteAt, req)
	return <-req.b
}

// A WriteByteAtReq holds the info needed to read a byte from a
// MemoryCommander.
type WriteByteAtReq struct {
	addr Word
	b    Byte
}

// WriteByteAt writes a single byte to the mmu at the specified address.
func (m *Mmu) WriteByteAt(addr Worder, b Byter) {
	req := WriteByteAtReq{addr.Word(), b.Byte()}
	m.RunCommand(CmdWriteByteAt, req)
}

func (m *Mmu) readByte(addr Worder) (Byte, error) {
	md, addr, err := m.selectMemoryDevice(addr)
	m.yield()
	return md.ReadByteAt(addr), err
}

func (m *Mmu) writeByte(addr Worder, b Byte) error {
	md, addr, err := m.selectMemoryDevice(addr)
	if err != nil {
		err = fmt.Errorf(fmt.Sprintf("%s - 0x%02X", err, b))
	}
	m.yield()
	md.WriteByteAt(addr, b)
	return err
}

// A RomDevice is a basic rom MemoryDevice.
type RomDevice struct {
	data []Byte
	addr Word
	size Word
}

// NewRomDevice creates a RomDevice that handles from `addr` to `addr + size`.
// The devices is initialized with `data` if provided.
func NewRomDevice(addr Worder, size Worder, data []Byte) RomDevice {
	d := make([]Byte, size.Word())
	copy(d, data)
	return RomDevice{d, addr.Word(), size.Word()}
}

// ReadByteAt reads a single byte from the device at the specified address.
func (r RomDevice) ReadByteAt(addr Worder) Byte {
	a := addr.Word() - r.addr
	if a < 0 || a > r.size {
		panic("rom read out of range")
	}
	return r.data[a]
}

// WriteByteAt does nothing.
func (r RomDevice) WriteByteAt(Worder, Byter) {}

// A RamDevice is a basic ram MemoryDevice.
type RamDevice struct {
	data []Byte
	addr Word
	size Word
}

// NewRamDevice creates a RamDevice that handles from `addr` to `addr + size`.
// The devices is initialized with `data` if provided.
func NewRamDevice(addr Worder, size Worder, data []Byte) RamDevice {
	d := make([]Byte, size.Word())
	copy(d, data)
	return RamDevice{d, addr.Word(), size.Word()}
}

// ReadByteAt reads a single byte from the device at the specified address.
func (r RamDevice) ReadByteAt(addr Worder) Byte {
	a := addr.Word() - r.addr
	if a < 0 || a > r.size {
		panic("ram read out of range")
	}
	return r.data[a]
}

// WriteByteAt writes a single byte to the device at the specified address.
func (r RamDevice) WriteByteAt(addr Worder, b Byter) {
	a := addr.Word() - r.addr
	if a < 0 || a > r.size {
		panic(fmt.Sprintf("ram write out of range 0x%04X 0x%02X", addr.Word(), b.Byte()))
	}
	r.data[a] = b.Byte()
}

// An EchoRamDevice is a ram MemoryDevice that can be accessed from two
// seperate starting addresses.
type EchoRamDevice struct {
	data  []Byte
	addrA Word
	addrB Word
	size  Word
}

// NewEchoRamDevice creates an EchoRamDevice that handles from `addr` to
// `addr + size`. The devices is initialized with `data` if provided.
func NewEchoRamDevice(addrA, addrB Worder, size Worder, data []Byte) EchoRamDevice {
	d := make([]Byte, size.Word())
	copy(d, data)
	return EchoRamDevice{d, addrA.Word(), addrB.Word(), size.Word()}
}

// ReadByteAt reads a single byte from the device at the specified address.
func (r EchoRamDevice) ReadByteAt(addr Worder) Byte {
	aa := addr.Word() - r.addrA
	ab := addr.Word() - r.addrB
	if aa < 0 || aa > r.size {
		return r.data[aa]
	} else if ab < 0 || ab > r.size {
		return r.data[ab]
	}
	panic("ram read out of range")
}

// WriteByteAt writes a single byte to the device at the specified address.
func (r EchoRamDevice) WriteByteAt(addr Worder, b Byter) {
	aa := addr.Word() - r.addrA
	ab := addr.Word() - r.addrB
	if aa < 0 || aa > r.size {
		r.data[aa] = b.Byte()
	} else if ab < 0 || ab > r.size {
		r.data[ab] = b.Byte()
	}
	panic("ram write out of range")
}

// A FunctionDevice is a MemoryDevice that maps custom read and write
// functions.
type FunctionDevice struct {
	fr func(Worder) Byte
	fw func(Worder, Byter)
}

// NewFunctionDevice creates a FunctionDevice that handles from `addr` to
// `addr + size`.
func NewFunctionDevice(fr func(Worder) Byte, fw func(Worder, Byter)) FunctionDevice {
	return FunctionDevice{fr, fw}
}

// ReadByteAt reads a single byte from the device at the specified address.
func (f FunctionDevice) ReadByteAt(addr Worder) Byte {
	return f.fr(addr)
}

// WriteByteAt writes a single byte to the device at the specified address.
func (f FunctionDevice) WriteByteAt(addr Worder, b Byter) {
	f.fw(addr, b)
}

// nil memory device
type nilModule struct{}

func (n nilModule) ReadByteAt(Worder) Byte {
	return Byte(0)
}

func (n nilModule) WriteByteAt(Worder, Byter) {}
