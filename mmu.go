package main

import (
	"fmt"
)

type mmu struct {
	bios memoryDevice // unloadable, maps to first 0xFF bytes when loaded
	rom  memoryDevice // 0000-7FFF 32k
	vram memoryDevice // 8000-9FFF 8k
	eram memoryDevice // A000-BFFF 8k
	wram memoryDevice // C000-DFFF 8k
	// echo ram F000-FDFF
	oam   memoryDevice // FE00-FE9F
	io    memoryDevice // FF00-FF40
	vidIo memoryDevice // FF40-FF49
	zero  memoryDevice // FF80-FFFF

	outBios *bool
}

type memoryDevice interface {
	readByte(addressInterface) uint8
	writeByte(addressInterface, uint8)
}

type ramModule []uint8

func newRamModule(size uint16, data []uint8) *ramModule {
	rm := make(ramModule, size)
	copy(rm, data)
	return &rm
}
func (r *ramModule) readByte(addr addressInterface) uint8 {
	a := addr.Uint16()
	if a > uint16(len(*r)) {
		panic("ram read out of range")
	}
	return (*r)[a]
}

func (r *ramModule) writeByte(addr addressInterface, b uint8) {
	a := addr.Uint16()
	if a > uint16(len(*r)) {
		panic("ram write out of range")
	}
	(*r)[a] = b
}

type romModule []uint8

func newRomModule(size uint16, data []uint8) *romModule {
	rm := make(romModule, size)
	copy(rm, data)
	return &rm
}
func (r *romModule) readByte(addr addressInterface) uint8 {
	a := addr.Uint16()
	if a > uint16(len(*r)) {
		panic("rom read out of range")
	}
	return (*r)[a]
}

func (r *romModule) writeByte(addressInterface, uint8) {
}

type nilModule struct{}

func (n nilModule) readByte(addr addressInterface) uint8 {
	return 0
}

func (r nilModule) writeByte(addressInterface, uint8) {
}

func newMmu(bios memoryDevice, cart cartridge, vid video) mmu {
	mc := mmu{
		bios:    bios,
		rom:     cart,
		vram:    vid,
		eram:    cart.eram,
		wram:    newRamModule(0x2000, nil),
		oam:     vid.oam,
		io:      newRamModule(0x4D, nil),
		vidIo:   vid.io,
		zero:    newRamModule(0x80, nil),
		outBios: new(bool)}
	return mc
}

type addressInterface interface {
	Uint16() uint16
}

type address uint16

func (u address) Uint16() uint16 { return uint16(u) }

func (mc mmu) String() string {
	return "<mmu>"
}

func (mc mmu) unloadBios() {
	*mc.outBios = true
}

// gets the memory device that handles an address
// returns the device and correct address
func (mc mmu) selectMemoryDevice(addr addressInterface) (memoryDevice, address) {
	a := addr.Uint16()
	if 0 <= a && a < 0xFF && !*mc.outBios {
		return mc.bios, address(a)
	} else if 0 <= a && a < 0x8000 {
		return mc.rom, address(a)
	} else if 0x8000 <= a && a < 0xA000 {
		return mc.vram, address(a - 0x8000)
	} else if 0xA000 <= a && a < 0xC000 { // switchable ram
		return mc.eram, address(a - 0xA000)
	} else if 0xC000 <= a && a < 0xE000 {
		return mc.wram, address(a - 0xC000)
	} else if 0xE000 <= a && a < 0xFE00 { // echo wram
		return mc.wram, address(a - 0xE000)
	} else if 0xFE00 <= a && a < 0xFEA0 {
		return mc.oam, address(a - 0xFE00)
		//} else if 0xFF00 <= a && a < 0xFF40 {
		//	return mc.io, address(a - 0xFF00)
	} else if 0xFF00 == a { // keypad
		panic("keypad unimplemented")
	} else if 0xFF01 <= a && a < 0xFF03 { // serial
		return mc.io, address(a - 0xFF00)
	} else if 0xFF04 == a { // DIV
		panic("div register unimplemented")
	} else if 0xFF0F == a { // IF
		return mc.io, address(a - 0xFF00)
	} else if 0xFF11 <= a && a < 0xFF15 { // NR11-14
		return mc.io, address(a - 0xFF00)
	} else if 0xFF24 <= a && a < 0xFF27 { // NR50-52
		return mc.io, address(a - 0xFF00)
	} else if 0xFF40 <= a && a < 0xFF49 {
		return mc.vidIo, address(a - 0xFF40)
	} else if 0xFF80 <= a && a <= 0xFFFF {
		return mc.zero, address(a - 0xFF80)
	}
	//return nilModule{}, address(0)
	panic(fmt.Sprintf("unhandled memory access: 0x%04X", addr))
}

func (mc mmu) readByte(addr addressInterface) uint8 {
	dev, a := mc.selectMemoryDevice(addr)
	return dev.readByte(a)
}

func (mc mmu) readWord(addr address) uint16 {
	l := mc.readByte(addr)
	h := mc.readByte(address(addr.Uint16() + 1))
	return bytesToWord(h, l)
}

func (mc mmu) writeByte(addr addressInterface, b uint8) {
	dev, a := mc.selectMemoryDevice(addr)
	dev.writeByte(a, b)
}

func (mc mmu) writeWord(addr address, w uint16) {
	h, l := wordToBytes(w)
	mc.writeByte(addr, l)
	mc.writeByte(address(addr.Uint16()+1), h)
}
