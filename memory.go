package main

import (
)

type memoryController struct {
	addr chan uint16
	data chan uint8

	rom  memoryDevice // 32k (only first four banks)
	vram memoryDevice // 8k
	eram memoryDevice // 8k (only first bank)
	wram memoryDevice // 8k
	iram memoryDevice // 8k
}

type memoryDevice interface {
	readByte(address) uint8
	writeByte(address, uint8)
}

// TODO: add support for banks
type ramModule []uint8

func newRamModule(size uint16, data []uint8) *ramModule {
	rm := make(ramModule, size)
	copy(rm, data)
	return &rm
}
func (r *ramModule) readByte(addr address) uint8 {
	a := addr.Uint16()
	if a > uint16(len(*r)) {
		panic("ram read out of range")
	}
	return (*r)[a]
}

func (r *ramModule) writeByte(addr address, b uint8) {
	a := addr.Uint16()
	if a > uint16(len(*r)) {
		panic("ram write out of range")
	}
	(*r)[a] = b
}

// TODO: add support for banks
type romModule []uint8

func newRomModule(size uint16, data []uint8) *romModule {
	rm := make(romModule, size)
	copy(rm, data)
	return &rm
}
func (r *romModule) readByte(addr address) uint8 {
	a := addr.Uint16()
	if a > uint16(len(*r)) {
		panic("rom read out of range")
	}
	return (*r)[a]
}

func (r *romModule) writeByte(address, uint8) {
	// nop
}

func newMemoryController(rom []uint8) memoryController {
	mc := memoryController{
		rom:  newRomModule(0x8000, rom),
		vram: newRamModule(0x2000, nil),
		eram: newRamModule(0x2000, nil),
		wram: newRamModule(0x2000, nil),
		iram: newRamModule(0x2000, nil)}
	return mc
}

type address interface {
	Uint16() uint16
}

type Uint16 uint16

func (u Uint16) Uint16() uint16 { return uint16(u) }

func (mc memoryController) readByte(addr address) uint8 {
	a := addr.Uint16()
	if a < 0x8000 {
		return mc.rom.readByte(Uint16(a))
	}
	a -= 0x8000
	if a < 0x2000 {
		return mc.vram.readByte(Uint16(a))
	}
	a -= 0x2000
	if a < 0x2000 {
		return mc.eram.readByte(Uint16(a))
	}
	a -= 0x2000
	if a < 0x2000 {
		return mc.wram.readByte(Uint16(a))
	}
	a -= 0x2000
	return mc.iram.readByte(Uint16(a))
}

func (mc *memoryController) readWord(addr address) uint16 {
	l := mc.readByte(addr)
	h := mc.readByte(Uint16(addr.Uint16() + 1))
	return uint16(h)<<8 + uint16(l)
}
