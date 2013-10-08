package jibi

import (
	"fmt"
)

// An Mmu is the memory management unit. Its purpose is to dispatch read and
// write requeststo the appropriate module (cpu, gpu, etc) based on the memory
// address. The Mmu is controlled by the cpu.
type Mmu struct {
	// memory
	mhandlers  []MemoryHandler
	cmhandlers []CpuMemoryHandler

	// internal state
	cpu    MemoryCommander
	rwChan chan Byte
}

// NewMmu creates a new Mmu with an optional bios that replaces 0x0000-0x00FF.
func NewMmu() *Mmu {
	// add ram only handlers
	workingRam := NewEchoRamDevice(Word(0xC000), Word(0xE000), Word(0x2000), nil)
	cmhandlers := []CpuMemoryHandler{
		CpuMemoryHandler{0xC000, 0xDFFF, workingRam},  // internal
		CpuMemoryHandler{0xE000, 0xFDFF, workingRam},  // echo
		CpuMemoryHandler{0xFEA0, 0xFEFF, nilDevice{}}, // unusable
		CpuMemoryHandler{0xFF50, 0xFF50, nilDevice{}}, // unusable, not sure why bios accesses it
		CpuMemoryHandler{0xFF7F, 0xFF7F, nilDevice{}}, // unusable, not sure why bios accesses it
	}
	mmu := &Mmu{
		nil,
		cmhandlers,
		nil,
		make(chan Byte, 1), // HACK
	}
	return mmu
}

// A LocalMemoryDevice provides random 16bit read and write access that will
// run on the cpu goroutine.
type LocalMemoryDevice interface {
	ReadLocalByteAt(Worder) Byte
	WriteLocalByteAt(Worder, Byter)
}

// A CpuMemoryHandler holds the info needed to map an address to a
// a CpuMemoryDevice.
type CpuMemoryHandler struct {
	start Word
	end   Word
	dev   LocalMemoryDevice
}

// A MemoryDevice provides random 16bit read and write access.
type MemoryDevice interface {
	ReadByteAt(Worder, chan Byte)
	WriteByteAt(Worder, Byter)
}

// A MemoryHandler holds the info needed to map an address to a MemoryDevice.
type MemoryHandler struct {
	start Word
	end   Word
	dev   MemoryDevice
}

func (m *Mmu) connectCpu(cpu MemoryCommander) {
	m.cpu = cpu
}

// incomplete, used for debugging
func (m *Mmu) getMemoryInfo(addr Worder) (string, bool) {
	a := addr.Word()
	if 0x9C00 <= a && a <= 0x9FFF {
		return "Background Map Data 2", false
	} else if a == 0xFF00 {
		return "Register for reading joy pad info and determining system type. (R/W)", false
	} else if a == 0xFF01 {
		return "Serial transfer data (R/W)", true
	} else if a == 0xFF02 {
		return "SIO control (R/W)", true
	} else if a == 0xFF10 {
		return "Sound Mode 1 register, Sweep register (R/W)", true
	} else if a == 0xFF11 {
		return "Sound Mode 1 register, Sound length/Wave pattern duty (R/W)", true
	} else if a == 0xFF12 {
		return "Sound Mode 1 register, Envelope (R/W)", true
	} else if a == 0xFF13 {
		return "Sound Mode 1 register, Frequency lo (W)", true
	} else if a == 0xFF14 {
		return "Sound Mode 1 register, Frequency hi (R/W)", true
	} else if a == 0xFF17 {
		return "Sound Mode 2 register, envelope (R/W)", true
	} else if a == 0xFF19 {
		return "Sound Mode 2 register, frequency", true
	} else if a == 0xFF1A {
		return "Sound Mode 3 register, Sound on/off (R/W)", true
	} else if a == 0xFF20 {
		return "Sound Mode 4 register, sound length (R/W)", true
	} else if a == 0xFF21 {
		return "Sound Mode 4 register, envelope (R/W)", true
	} else if a == 0xFF23 {
		return "Sound Mode 4 register, counter/consecutive; inital (R/W)", true
	} else if a == 0xFF24 {
		return "Channel control / ON-OFF / Volume (R/W)", true
	} else if a == 0xFF25 {
		return "Selection of Sound output terminal (R/W)", true
	} else if a == 0xFF26 {
		return "Sound on/off (R/W)", true
	}
	return "unknown", false

}

func (m *Mmu) selectLocalMemoryDevice(addr Worder) LocalMemoryDevice {
	a := addr.Word()
	for _, mh := range m.cmhandlers {
		if mh.start <= a && a <= mh.end {
			return mh.dev
		}
	}
	u, v := m.getMemoryInfo(addr)
	if !v {
		panic(fmt.Sprintf("unhandled memory access: 0x%04X - %s", a, u))
	}
	return nilDevice{}
}

func (m *Mmu) selectMemoryDevice(addr Worder) MemoryDevice {
	a := addr.Word()
	for _, mh := range m.mhandlers {
		if mh.start <= a && a <= mh.end {
			return mh.dev
		}
	}
	u, v := m.getMemoryInfo(addr)
	if !v {
		panic(fmt.Sprintf("unhandled memory access: 0x%04X - %s", a, u))
	}
	return nilDevice{}
}

// A ReadByteAtReq holds the info needed to read a byte from a MemoryCommander.
type ReadByteAtReq struct {
	addr Word
	b    chan Byte
}

// ReadByteAt reads a single byte from the mmu at the specified address into
// the provided channel.
func (m *Mmu) ReadByteAt(addr Worder, b chan Byte) {
	m.cpu.ReadByteAt(addr, b)
}

// A WriteByteAtReq holds the info needed to read a byte from a
// MemoryCommander.
type WriteByteAtReq struct {
	addr Word
	b    Byte
}

// WriteByteAt writes a single byte to the mmu at the specified address.
func (m *Mmu) WriteByteAt(addr Worder, b Byter) {
	m.cpu.WriteByteAt(addr, b)
}

func (m *Mmu) isLocalMemory(addr Worder) bool {
	for _, mh := range m.cmhandlers {
		if mh.start <= addr.Word() && addr.Word() <= mh.end {
			return true
		}
	}
	return false
}

func (m *Mmu) readLocalByte(addr Worder) Byte {
	md := m.selectLocalMemoryDevice(addr)
	return md.ReadLocalByteAt(addr)
}

func (m *Mmu) writeLocalByte(addr Worder, b Byter) {
	md := m.selectLocalMemoryDevice(addr)
	md.WriteLocalByteAt(addr, b)
}

func (m *Mmu) readRemoteByte(addr Worder) Byte {
	md := m.selectMemoryDevice(addr)
	md.ReadByteAt(addr, m.rwChan)
	return <-m.rwChan
}

func (m *Mmu) writeRemoteByte(addr Worder, b Byter) {
	md := m.selectMemoryDevice(addr)
	md.WriteByteAt(addr, b)
}

// HandleMemory maps a start and end address to a MemoryDevice.
func (m *Mmu) HandleMemory(start, end Word, md MemoryDevice) {
	m.cpu.RunCommand(CmdHandleMemory, MemoryHandler{start, end, md})
}

// HandleCpuMemory maps a start and end address to a LocalMemoryDevice. This
// device's memory access is run on the cpu goroutine.
func (m *Mmu) HandleCpuMemory(start, end Word, md LocalMemoryDevice) {
	m.cpu.RunCommand(CmdHandleCpuMemory, CpuMemoryHandler{start, end, md})
}

func (m *Mmu) cmdHandleCpuMemory(resp interface{}) {
	if v, ok := resp.(CpuMemoryHandler); !ok {
		panic("invalid command response type")
	} else {
		m.handleLocalMemory(v)
	}
}

func (m *Mmu) cmdHandleMemory(resp interface{}) {
	if v, ok := resp.(MemoryHandler); !ok {
		panic("invalid command response type")
	} else {
		m.handleMemory(v)
	}
}

func (m *Mmu) handleMemory(mh MemoryHandler) {
	m.mhandlers = append(m.mhandlers, mh)
}

func (m *Mmu) handleLocalMemory(mh CpuMemoryHandler) {
	m.cmhandlers = append(m.cmhandlers, mh)
}

// A RomDevice is a basic rom MemoryDevice.
type RomDevice struct {
	data []Byte
	addr Word
	size Word
}

// NewRomDevice creates a RomDevice that handles from `addr` to `addr + size`.
// The devices is initialized with `data` if provided. All operations run on
// the callers goroutine.
func NewRomDevice(addr Worder, size Worder, data []Byte) RomDevice {
	d := make([]Byte, size.Word())
	copy(d, data)
	return RomDevice{d, addr.Word(), size.Word()}
}

// ReadLocalByteAt reads a single byte from the device at the specified address.
func (r RomDevice) ReadLocalByteAt(addr Worder) Byte {
	a := addr.Word() - r.addr
	if a < 0 || a > r.size {
		panic("rom read out of range")
	}
	return r.data[a]
}

// WriteLocalByteAt does nothing.
func (r RomDevice) WriteLocalByteAt(Worder, Byter) {}

// A RamDevice is a basic ram MemoryDevice.
type RamDevice struct {
	data []Byte
	addr Word
	size Word
}

// NewRamDevice creates a RamDevice that handles from `addr` to `addr + size`.
// The devices is initialized with `data` if provided. All operations run on
// the callers goroutine.
func NewRamDevice(addr Worder, size Worder, data []Byte) RamDevice {
	d := make([]Byte, size.Word())
	copy(d, data)
	return RamDevice{d, addr.Word(), size.Word()}
}

// ReadByteAt reads a single byte from the device at the specified address.
func (r RamDevice) ReadByteAt(addr Worder, b chan Byte) {
	a := addr.Word() - r.addr
	if a < 0 || a > r.size {
		panic("ram read out of range")
	}
	b <- r.data[a]
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

// ReadLocalByteAt reads a single byte from the device at the specified address.
func (r EchoRamDevice) ReadLocalByteAt(addr Worder) Byte {
	aa := addr.Word() - r.addrA
	ab := addr.Word() - r.addrB
	if aa < r.size {
		return r.data[aa]
	} else if ab < r.size {
		return r.data[ab]
	}
	panic("echo ram read out of range")
}

// WriteLocalByteAt writes a single byte to the device at the specified address.
func (r EchoRamDevice) WriteLocalByteAt(addr Worder, b Byter) {
	aa := addr.Word() - r.addrA
	ab := addr.Word() - r.addrB
	if aa < r.size {
		r.data[aa] = b.Byte()
	} else if ab < r.size {
		r.data[ab] = b.Byte()
	} else {
		panic("echo ram write out of range")
	}
}

// nil memory device
type nilDevice struct{}

func (n nilDevice) ReadLocalByteAt(Worder) Byte {
	return Byte(0)
}

func (n nilDevice) ReadByteAt(addr Worder, b chan Byte) {
	b <- Byte(0)
}

func (n nilDevice) WriteLocalByteAt(Worder, Byter) {}

func (n nilDevice) WriteByteAt(Worder, Byter) {}
