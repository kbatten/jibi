package jibi

import (
	"fmt"
	"sync"
)

// A list of all the special memory addresses.
const (
	AddrRom    Word = 0x0000
	AddrVRam   Word = 0x8000
	AddrERam   Word = 0xA000
	AddrRam    Word = 0xC000
	AddrOam    Word = 0xFE00
	AddrOamEnd Word = 0xFEA0

	AddrP1   Word = 0xFF00
	AddrDIV  Word = 0xFF04
	AddrTIMA Word = 0xFF05
	AddrTMA  Word = 0xFF06
	AddrTAC  Word = 0xFF07
	AddrIF   Word = 0xFF0F

	AddrGpuRegs    Word = 0xFF40
	AddrLCDC       Word = 0xFF40
	AddrSTAT       Word = 0xFF41
	AddrSCY        Word = 0xFF42
	AddrSCX        Word = 0xFF43
	AddrLY         Word = 0xFF44
	AddrLYC        Word = 0xFF45
	AddrDMA        Word = 0xFF46
	AddrBGP        Word = 0xFF47
	AddrOBP0       Word = 0xFF48
	AddrOBP1       Word = 0xFF49
	AddrWY         Word = 0xFF4A
	AddrWX         Word = 0xFF4B
	AddrGpuRegsEnd Word = 0xFF4C

	AddrZero Word = 0xFF80
	AddrIE   Word = 0xFFFF
)

// An Mmu is the memory management unit. Its purpose is to dispatch read and
// write requeststo the appropriate module (cpu, gpu, etc) based on the memory
// address. The Mmu is controlled by the cpu.
type Mmu interface {
	ReadByteAt(addr Word) Byte
	WriteByteAt(addr Word, b Byte)
	WriteElevatedByteAt(addr Word, b Byte)
	ReadWordAt(addr Word) Word
	WriteWordAt(addr, w Word)
	ReadIoByte(addr Word) Byte
	SetKeypad(kp *Keypad)
	SetGpu(gpu *Gpu)
	SetInterrupt(in Interrupt)
	ResetInterrupt(in Interrupt)
}

type RomOnlyMmu struct {
	// memory blocks and io
	rom     []Byte
	vram    []Byte
	ram     []Byte
	oam     []Byte
	ioP1    *mmio
	div     Byte
	tima    Byte
	tma     Byte
	tac     Byte
	ioIF    *mmio
	gpuregs []Byte
	zero    []Byte
	ie      Byte

	// memory locks
	onlyMemoryLock *sync.Mutex

	// internal state
	kp  *Keypad
	gpu *Gpu
}

// NewMmu creates a new Mmu with an optional bios that replaces 0x0000-0x00FF.
func NewMmu(cart *Cartridge) Mmu {
	var rom []Byte
	if cart != nil {
		rom = cart.Rom
	}
	mmu := &RomOnlyMmu{
		rom:            rom,
		vram:           make([]Byte, 0x2000),
		ram:            make([]Byte, 0x2000),
		oam:            make([]Byte, 0xA0),
		ioP1:           newMmio(AddrP1),
		div:            Byte(0),
		tima:           Byte(0),
		tma:            Byte(0),
		tac:            Byte(0),
		ioIF:           newMmio(AddrIF),
		gpuregs:        make([]Byte, 12),
		zero:           make([]Byte, 0x100),
		onlyMemoryLock: new(sync.Mutex),
	}
	return mmu
}

type addressBlock uint16

const (
	abNil addressBlock = iota
	abRom addressBlock = 1 << iota
	abVRam
	abERam
	abRam
	abOam
	abP1
	abDIV
	abTIMA
	abTMA
	abTAC
	abIF
	abGpuRegs
	abZero
	abIE
	abLast = abIE
)

func (a addressBlock) String() string {
	switch a {
	case abNil:
		return "abNil"
	case abRom:
		return "abRom"
	case abVRam:
		return "abVRam"
	case abERam:
		return "abERam"
	case abRam:
		return "abRam"
	case abOam:
		return "abOam"
	case abIF:
		return "abIF"
	case abGpuRegs:
		return "abGpuRegs"
	case abZero:
		return "abZero"
	case abIE:
		return "abIE"
	}
	return "abUNKNOWN"
}

func (m *RomOnlyMmu) SetKeypad(kp *Keypad) {
	m.kp = kp
}

func (m *RomOnlyMmu) SetGpu(gpu *Gpu) {
	m.gpu = gpu
}

func (m *RomOnlyMmu) selectAddressBlock(addr Word, rw string) (addressBlock, Word) {
	if addr < AddrVRam {
		return abRom, 0
	} else if AddrVRam <= addr && addr < AddrERam {
		return abVRam, AddrVRam
	} else if AddrERam <= addr && addr < AddrRam {
		return abERam, AddrERam
	} else if AddrRam <= addr && addr < AddrOam {
		return abRam, AddrRam
	} else if AddrOam <= addr && addr < AddrOamEnd {
		return abOam, AddrOam
	} else if AddrP1 == addr {
		return abP1, AddrP1
	} else if AddrDIV == addr {
		return abDIV, AddrDIV
	} else if AddrTIMA == addr {
		return abTIMA, AddrTIMA
	} else if AddrTMA == addr {
		return abTMA, AddrTMA
	} else if AddrTAC == addr {
		return abTAC, AddrTAC
	} else if AddrIF == addr {
		return abIF, AddrIF
	} else if AddrGpuRegs <= addr && addr < AddrGpuRegsEnd {
		return abGpuRegs, AddrGpuRegs
	} else if AddrZero <= addr && addr < AddrIE {
		return abZero, AddrZero
	} else if AddrIE == addr {
		return abIE, AddrIE
	}

	u, v := m.getAddressInfo(addr)
	if !v {
		if rw == "" {
			rw = "access"
		}
		panic(fmt.Sprintf("unhandled memory %s: 0x%04X - %s", rw, addr, u))
	}
	return abNil, 0
}

func (m *RomOnlyMmu) ReadByteAt(addr Word) Byte {
	m.onlyMemoryLock.Lock()
	defer m.onlyMemoryLock.Unlock()

	blk, start := m.selectAddressBlock(addr, "read")
	if blk == abRom {
		return m.rom[addr-start]
	}
	if blk == abVRam {
		return m.vram[addr-start]
	} else if blk == abRam {
		return m.ram[(addr-start)&0x1FFF]
	} else if blk == abOam {
		return m.oam[addr-start]
	} else if blk == abP1 {
		return m.ioP1.readByte()
	} else if blk == abDIV {
		return m.div
	} else if blk == abTIMA {
		return m.tima
	} else if blk == abTMA {
		return m.tma
	} else if blk == abTAC {
		return m.tac
	} else if blk == abIF {
		return m.ioIF.readByte()
	} else if blk == abGpuRegs {
		return m.gpuregs[addr-start]
	} else if blk == abZero {
		return m.zero[addr-start]
	} else if blk == abIE {
		return m.ie
	}
	if u, v := m.getAddressInfo(addr); !v {
		panic(fmt.Sprintf("unhandled memory read: 0x%04X - %s", addr, u))
	}
	return 0
}

func (m *RomOnlyMmu) WriteByteAt(addr Word, b Byte) {
	m.onlyMemoryLock.Lock()
	defer m.onlyMemoryLock.Unlock()

	blk, start := m.selectAddressBlock(addr, "write")
	if blk == abRom {
		return
	} else if blk == abVRam {
		m.vram[addr-start] = b
		return
	} else if blk == abRam {
		m.ram[(addr-start)&0x1FFF] = b
		return
	} else if blk == abOam {
		m.oam[addr-start] = b
		return
	} else if blk == abP1 {
		m.ioP1.writeByte(b)
		return
	} else if blk == abDIV {
		m.div = Byte(0) // writing any value sets to it 0
		return
	} else if blk == abTIMA {
		m.tima = b
		return
	} else if blk == abTMA {
		m.tma = b
		return
	} else if blk == abTAC {
		m.tac = b
		return
	} else if blk == abIF {
		m.ioIF.writeByte(b)
		return
	} else if blk == abGpuRegs {
		if addr == AddrLCDC {
			prevBit7 := m.gpuregs[addr-start] & 0x80
			bit7 := b & 0x80
			if prevBit7 == 0 && bit7 != 0 {
				panic("gpu run")
				m.gpu.RunCommand(CmdPlay, nil)
			} else if prevBit7 != 0 && bit7 == 0 {
				panic("gpu pause")
				m.gpu.RunCommand(CmdPause, nil)
				m.gpuregs[AddrLY-start] = 0
			}
		} else if addr == AddrLY {
			m.gpuregs[addr-start] = 0 // reset on write
		} else {
			m.gpuregs[addr-start] = b
		}
		return
	} else if blk == abZero {
		m.zero[addr-start] = b
		return
	} else if blk == abIE {
		m.ie = b
		return
	}
	if u, v := m.getAddressInfo(addr); !v {
		panic(fmt.Sprintf("unhandled memory write: 0x%04X - %s", addr, u))
	}
}

// for memory that always writes 0 on write, write the value instead
func (m *RomOnlyMmu) WriteElevatedByteAt(addr Word, b Byte) {
	m.onlyMemoryLock.Lock()
	defer m.onlyMemoryLock.Unlock()

	blk, start := m.selectAddressBlock(addr, "write")
	if blk == abDIV {
		m.div = b
		return
	} else if blk == abGpuRegs {
		if addr == AddrLY {
			m.gpuregs[addr-start] = b
			return
		}
	}
	if u, v := m.getAddressInfo(addr); !v {
		panic(fmt.Sprintf("unhandled memory write: 0x%04X - %s", addr, u))
	}
}

func (m *RomOnlyMmu) ReadWordAt(addr Word) Word {
	return BytesToWord(m.ReadByteAt(addr+1), m.ReadByteAt(addr))
}

func (m *RomOnlyMmu) WriteWordAt(addr, w Word) {
	m.WriteByteAt(addr+1, w.High())
	m.WriteByteAt(addr, w.Low())
}

func (m *RomOnlyMmu) ReadIoByte(addr Word) Byte {
	blk, _ := m.selectAddressBlock(addr, "write")
	if blk == abP1 {
		return m.ioP1.readByte()
	} else if blk == abIF {
		return m.ioIF.readByte()
	}
	panic(fmt.Sprintf("unhandled queued write: 0x%04X", addr))
}

// incomplete, used for debugging
// a return value of true means we can ignore this address
func (m *RomOnlyMmu) getAddressInfo(addr Word) (string, bool) {
	if 0x9C00 <= addr && addr <= 0x9FFF {
		return "Background Map Data 2", false
	} else if 0xA000 <= addr && addr <= 0xBFFF {
		return "Cartridge Ram", true // TODO: find out what should happen in rom only
	} else if 0xFEA0 <= addr && addr <= 0xFEFF {
		return "unusable memory", true
	} else if addr == 0xFF00 {
		return "Register for reading joy pad info and determining system type. (R/W)", false
	} else if addr == 0xFF01 {
		return "Serial transfer data (R/W)", true
	} else if addr == 0xFF02 {
		return "SIO control (R/W)", true
	} else if addr == 0xFF03 {
		return "no clue", true
	} else if addr == 0xFF04 {
		return "DIV", false
	} else if addr == 0xFF05 {
		return "TIMA", false
	} else if addr == 0xFF06 {
		return "TMA", false
	} else if addr == 0xFF07 {
		return "TAC", false
	} else if 0xFF08 <= addr && addr <= 0xFF0E {
		return "no clue", true
	} else if addr == 0xFF10 {
		return "Sound Mode 1 register, Sweep register (R/W)", true
	} else if addr == 0xFF11 {
		return "Sound Mode 1 register, Sound length/Wave pattern duty (R/W)", true
	} else if addr == 0xFF12 {
		return "Sound Mode 1 register, Envelope (R/W)", true
	} else if addr == 0xFF13 {
		return "Sound Mode 1 register, Frequency lo (W)", true
	} else if addr == 0xFF14 {
		return "Sound Mode 1 register, Frequency hi (R/W)", true
	} else if addr == 0xFF15 {
		return "no clue", true
	} else if addr == 0xFF16 {
		return "Sound Mode 2 register, Sound Length; Wave Pattern Duty (R/W)", true
	} else if addr == 0xFF17 {
		return "Sound Mode 2 register, envelope (R/W)", true
	} else if addr == 0xFF18 {
		return "Sound Mode 2 register, frequency lo data (W)", true
	} else if addr == 0xFF19 {
		return "Sound Mode 2 register, frequency", true
	} else if addr == 0xFF1A {
		return "Sound Mode 3 register, Sound on/off (R/W)", true
	} else if addr == 0xFF1B {
		return "Sound Mode 3 register, sound length (R/W)", true
	} else if addr == 0xFF1C {
		return "Sound Mode 3 register, Select output level (R/W)", true
	} else if addr == 0xFF1D {
		return "Sound Mode 3 register, frequency's lower data (W)", true
	} else if addr == 0xFF1E {
		return "Sound Mode 3 register, frequency's higher data (R/W)", true
	} else if addr == 0xFF1F {
		return "no clue", true
	} else if addr == 0xFF20 {
		return "Sound Mode 4 register, sound length (R/W)", true
	} else if addr == 0xFF21 {
		return "Sound Mode 4 register, envelope (R/W)", true
	} else if addr == 0xFF22 {
		return "Sound Mode 4 register, polynomial counter (R/W)", true
	} else if addr == 0xFF23 {
		return "Sound Mode 4 register, counter/consecutive; inital (R/W)", true
	} else if addr == 0xFF24 {
		return "Channel control / ON-OFF / Volume (R/W)", true
	} else if addr == 0xFF25 {
		return "Selection of Sound output terminal (R/W)", true
	} else if addr == 0xFF26 {
		return "Sound on/off (R/W)", true
	} else if 0xFF27 <= addr && addr <= 0xFF2F {
		return "no clue", true
	} else if 0xFF30 <= addr && addr <= 0xFF3F {
		return "Sound Sample RAM", true
	} else if addr == 0xFF47 {
		return "BGP", false
	} else if 0xFF4C == addr {
		return "no clue", true
	} else if 0xFF4D <= addr && addr <= 0xFF7F {
		return "GBC", true
	} else if addr == 0xFFFF {
		return "IE", false
	}
	return "unknown", false
}

// memory mapped io
type mmio struct {
	addr Word

	value Byte

	lock *sync.Mutex
}

func newMmio(addr Word) *mmio {
	m := &mmio{addr: addr,
		lock: new(sync.Mutex)}
	return m
}

func (m *mmio) readByte() Byte {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.value
}

func (m *mmio) writeByte(b Byte) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.value = b
}

func (mmu *RomOnlyMmu) SetInterrupt(in Interrupt) {
	iflags := mmu.ReadByteAt(AddrIF)
	mmu.WriteByteAt(AddrIF, iflags|Byte(in))
}

func (mmu *RomOnlyMmu) ResetInterrupt(in Interrupt) {
	iflags := mmu.ReadByteAt(AddrIF)
	mmu.WriteByteAt(AddrIF, iflags&(Byte(in)^0xFF))
}
